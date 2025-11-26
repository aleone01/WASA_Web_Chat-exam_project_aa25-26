package api

import 
(
	"encoding/json"
	"net/http"
)

// GetMyConversations (GET /users/{userId}/chats)
func (rt *Router) GetMyConversations(w http.ResponseWriter, r *http.Request, userId int, params GetMyConversationsParams) 
{
	
	// Check Auth
	if err := rt.checkAuth(w, r, userId); err != nil 
	{
		return
	}

	// chiamata al database per le chat

	// controlla se c'è il parametro ?username=...
	if params.Username != nil && *params.Username != "" 
	{
		// ricerca specifica
		chats, err := rt.db.GetChatWithUser(userId, *params.Username)
		if err != nil 
		{
			rt.sendError(w, http.StatusInternalServerError, 500, "Errore ricerca chat")
			return
		}
		rt.sendJSON(w, chats) // Helper ipotetico o usa json.NewEncoder ???
		return
	}

	// recupera tutte le chat
	chats, err := rt.db.GetConversations(userId)
	if err != nil 
	{
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore recupero chat")
		return
	}

	// risposta con le chat 
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {Chats interface{} `json:"chats"`}{Chats: chats})

}

// GetConversation (GET /users/{userId}/chats/{chatId}) 
func (rt *Router) GetConversation(w http.ResponseWriter, r *http.Request, userId int, chatId int) 
{
	
	// Check Auth
	if err := rt.checkAuth(w, r, userId); err != nil 
	{
		return
	}

	// chiamata al database per i messaggi
	messages, err := rt.db.GetChatMessages(chatId, userId)
	if err != nil 
	{
		// in caso di accesso negato
		if err.Error() == "Chat non trovata o accesso negato" 
		{
			rt.sendError(w, http.StatusForbidden, 403, "Accesso negato")
			return
		}
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore server")
		return
	}

	// risposta con i messaggi
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {Messages interface{} `json:"messages"`}{Messages: messages})

}
