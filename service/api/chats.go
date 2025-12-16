package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/aleone01/Web-Project-repo/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// createConversation inizializza una nuova chat privata (1-a-1) tra l'utente richiedente e un altro utente specificato.
// La funzione verifica l'autenticazione, impedisce la creazione di chat con se stessi e controlla se esiste già
// una conversazione attiva tra i due utenti. Se la chat esiste, la restituisce invece di crearne una duplicata.
func (rt *_router) createConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	userId, _ := strconv.Atoi(ps.ByName("userId"))
	if !rt.checkAuth(w, r, userId) {
		return
	}

	var reqBody struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido")
		return
	}

	otherUser, err := rt.db.GetUserByUsername(reqBody.Username)
	if err != nil {
		rt.sendError(w, http.StatusNotFound, 404, "Utente non trovato")
		return
	}
	if otherUser.Id == userId {
		rt.sendError(w, http.StatusBadRequest, 400, "Non puoi chattare con te stesso")
		return
	}

	existingChats, err := rt.db.GetChatWithUser(userId, reqBody.Username)
	if err == nil && len(existingChats) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(existingChats[0])
		return
	}

	newChat, err := rt.db.CreateConversation(userId, otherUser.Id)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore creazione chat")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(newChat)
}

// getMyConversations recupera la lista delle conversazioni a cui partecipa l'utente.
// Supporta un parametro di query opzionale 'username' per filtrare la ricerca e restituire solo la chat
// con uno specifico utente, altrimenti restituisce l'elenco completo delle conversazioni.
func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	userId, _ := strconv.Atoi(ps.ByName("userId"))

	if !rt.checkAuth(w, r, userId) {
		return
	}

	targetUsername := r.URL.Query().Get("username")
	if targetUsername != "" {
		chats, err := rt.db.GetChatWithUser(userId, targetUsername)
		if err != nil {
			rt.sendError(w, http.StatusInternalServerError, 500, "Errore server")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(struct {
			Chats interface{} `json:"chats"`
		}{chats})
		return
	}

	chats, err := rt.db.GetMyConversations(userId)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore server")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		Chats interface{} `json:"chats"`
	}{chats})
}

// getConversation restituisce il dettaglio di una specifica conversazione, inclusi i messaggi scambiati.
// Verifica che l'utente richiedente sia autorizzato ad accedere ai dati della chat richiesta.
func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	userId, _ := strconv.Atoi(ps.ByName("userId"))

	chatId, _ := strconv.Atoi(ps.ByName("chatId"))
	if !rt.checkAuth(w, r, userId) {
		return
	}

	msgs, err := rt.db.GetConversation(userId, chatId)

	if err != nil {
		rt.sendError(w, http.StatusForbidden, 403, "Errore DB")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		Messages interface{} `json:"messages"`
	}{msgs})
}
