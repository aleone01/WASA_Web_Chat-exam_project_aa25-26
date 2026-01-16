package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/aleone01/Web-Project-repo/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// createConversation inizializza una nuova chat privata (1-a-1) tra l'utente richiedente e un altro utente specificato.
// La funzione segue questi step:
// 1. Verifica autenticazione.
// 2. Controllo esistenza utente target e prevenzione auto-chat.
// 3. Controllo idempotenza: se esiste già una conversazione attiva, restituisce quella esistente (200 OK) invece di crearne una nuova.
// 4. Creazione nuova conversazione se non esiste (201 Created).
func (rt *_router) createConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	userId, _ := strconv.Atoi(ps.ByName("userId"))
	// Controllo autorizzativo standard
	if !rt.checkAuth(w, r, userId) {
		return
	}

	// Parsing del body per ottenere l'username del destinatario
	var reqBody struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido")
		return
	}

	// Verifica esistenza dell'altro utente nel sistema
	otherUser, err := rt.db.GetUserByUsername(reqBody.Username)
	if err != nil {
		rt.sendError(w, http.StatusNotFound, 404, "Utente destinatario non trovato")
		return
	}

	// Prevenzione creazione chat con se stessi
	if otherUser.Id == userId {
		rt.sendError(w, http.StatusBadRequest, 400, "Non puoi avviare una chat con te stesso")
		return
	}

	// Verifica se esiste già una chat tra i due utenti (Idempotenza)
	existingChats, err := rt.db.GetChatWithUser(userId, reqBody.Username)
	if err == nil && len(existingChats) > 0 {
		// Chat trovata, restituisco quella esistente
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(existingChats[0])
		return
	}

	// Creazione di una nuova entry nel DB
	newChat, err := rt.db.CreateConversation(userId, otherUser.Id)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore server durante la creazione della chat")
		return
	}

	// Restituzione della nuova chat
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(newChat)
}

// getMyConversations recupera la lista delle conversazioni a cui partecipa l'utente.
// Questa funzione opera in due modalità:
// 1. Filtro per utente: se è presente il parametro query 'username', cerca la chat specifica con quell'utente.
// 2. Lista completa: altrimenti restituisce tutte le conversazioni attive dell'utente.
func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	userId, _ := strconv.Atoi(ps.ByName("userId"))

	if !rt.checkAuth(w, r, userId) {
		return
	}

	// Controllo parametri Query String (?username=...)
	targetUsername := r.URL.Query().Get("username")
	if targetUsername != "" {
		// Modalità filtrata
		chats, err := rt.db.GetChatWithUser(userId, targetUsername)
		if err != nil {
			rt.sendError(w, http.StatusInternalServerError, 500, "Errore server nella ricerca chat")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(struct {
			Chats interface{} `json:"chats"`
		}{chats})
		return
	}

	// Modalità lista completa
	chats, err := rt.db.GetMyConversations(userId)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore server nel recupero lista conversazioni")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		Chats interface{} `json:"chats"`
	}{chats})
}

// getConversation restituisce il dettaglio completo (storico messaggi) di una specifica conversazione.
// Verifica che l'utente richiedente sia un partecipante autorizzato della chat.
func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	userId, _ := strconv.Atoi(ps.ByName("userId"))
	chatId, _ := strconv.Atoi(ps.ByName("chatId"))

	if !rt.checkAuth(w, r, userId) {
		return
	}

	// Recupero messaggi dal DB. La funzione GetConversation del DB dovrebbe gestire internamente
	// o restituire errore se l'utente non fa parte della chat.
	msgs, err := rt.db.GetConversation(userId, chatId)
	if err != nil {
		rt.sendError(w, http.StatusForbidden, 403, "Errore DB o Accesso Negato alla conversazione")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		Messages interface{} `json:"messages"`
	}{msgs})
}
