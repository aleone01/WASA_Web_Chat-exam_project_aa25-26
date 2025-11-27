package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/aleone01/Web-Project-repo/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getMyConversations gestisce il recupero delle conversazioni dell'utente
func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// estrazione userId dai parametri della rotta
	userId, _ := strconv.Atoi(ps.ByName("userId"))

	// verifica autenticazione
	if !rt.checkAuth(w, r, userId) {
		return
	}

	// controllo se è presente il filtro per username
	targetUsername := r.URL.Query().Get("username")
	if targetUsername != "" {
		chats, err := rt.db.GetChatWithUser(userId, targetUsername)
		if err != nil {
			rt.sendError(w, http.StatusInternalServerError, 500, "Errore server")
			return
		}

		// risposta con la chat trovata
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(struct {
			Chats interface{} `json:"chats"`
		}{chats})
		return
	}

	// recupero di tutte le conversazioni dell'utente
	chats, err := rt.db.GetConversations(userId)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore server")
		return
	}

	// risposta con le conversazioni
	w.Header().Set("Content-Type", "application/json")
	// CORREZIONE 2: Assegnazione a _
	_ = json.NewEncoder(w).Encode(struct {
		Chats interface{} `json:"chats"`
	}{chats})
}

// getConversation gestisce il recupero dei messaggi di una conversazione
func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// estrazione userId e chatId dai parametri della rotta
	userId, _ := strconv.Atoi(ps.ByName("userId"))
	chatId, _ := strconv.Atoi(ps.ByName("chatId"))

	// verifica autenticazione
	if !rt.checkAuth(w, r, userId) {
		return
	}

	// recupero dei messaggi della chat dal database
	msgs, err := rt.db.GetChatMessages(chatId, userId)
	if err != nil {
		rt.sendError(w, http.StatusForbidden, 403, "Accesso negato o chat inesistente")
		return
	}

	// risposta con i messaggi della chat
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		Messages interface{} `json:"messages"`
	}{msgs})
}
