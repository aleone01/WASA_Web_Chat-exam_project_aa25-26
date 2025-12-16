package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/aleone01/Web-Project-repo/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// sendMessage gestisce l'invio di un nuovo messaggio all'interno di una specifica chat.
// La funzione verifica l'autenticazione dell'utente, decodifica il payload JSON (testo, foto, riferimenti di risposta o inoltro)
// e registra il messaggio nel database con il timestamp corrente.
func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	userId, _ := strconv.Atoi(ps.ByName("userId"))
	chatId, _ := strconv.Atoi(ps.ByName("chatId"))

	if !rt.checkAuth(w, r, userId) {
		return
	}

	var reqBody struct {
		Text      string `json:"text"`
		Photo     string `json:"photoUrl"`
		ReplyTo   int    `json:"replyTo"`
		IsForward bool   `json:"isForward"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido")
		return
	}

	sentAt := time.Now()

	msg, err := rt.db.CreateMessage(chatId, userId, reqBody.Text, reqBody.Photo, sentAt, reqBody.ReplyTo, reqBody.IsForward)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore nel salvataggio")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(msg)
}

// commentMessage consente a un utente di aggiungere una reazione (emoticon) a un messaggio esistente.
// Identifica il messaggio e l'utente dai parametri e dal token, quindi aggiorna lo stato del messaggio nel database
// aggiungendo la reazione specificata.
func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	userId, _ := strconv.Atoi(ps.ByName("userId"))
	messageId, _ := strconv.Atoi(ps.ByName("messageId"))

	if !rt.checkAuth(w, r, userId) {
		return
	}

	var reqBody struct {
		Emoticon string `json:"emoticon"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido")
		return
	}

	if err := rt.db.AddReaction(messageId, userId, reqBody.Emoticon); err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore nell'aggiunta della reazione")
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// uncommentMessage permette a un utente di rimuovere una reazione precedentemente aggiunta a un messaggio.
// Esegue l'operazione inversa di commentMessage, eliminando l'associazione tra l'utente, il messaggio e l'emoticon nel database.
func (rt *_router) uncommentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	userId, _ := strconv.Atoi(ps.ByName("userId"))
	messageId, _ := strconv.Atoi(ps.ByName("messageId"))

	if !rt.checkAuth(w, r, userId) {
		return
	}

	if err := rt.db.RemoveReaction(messageId, userId); err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore nella rimozione reazione")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// deleteMessage gestisce la cancellazione di un messaggio inviato.
// La funzione verifica rigorosamente che l'utente richiedente sia l'autore del messaggio prima di procedere
// con l'eliminazione dal database. Restituisce un errore 403 Forbidden se si tenta di cancellare messaggi altrui.
func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	userId, _ := strconv.Atoi(ps.ByName("userId"))
	messageId, _ := strconv.Atoi(ps.ByName("messageId"))

	if !rt.checkAuth(w, r, userId) {
		return
	}

	if err := rt.db.DeleteMessage(messageId, userId); err != nil {
		if err.Error() == "message not found or unauthorized" {
			rt.sendError(w, http.StatusForbidden, 403, "Non autorizzato a cancellare questo messaggio")
			return
		}
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore nella cancellazione")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// forwardMessage gestisce l'inoltro di un messaggio esistente verso una o più chat di destinazione.
// Recupera il contenuto del messaggio originale e crea nuove istanze di messaggio per ogni destinatario specificato
// nel corpo della richiesta, marcandoli come "inoltrati".
func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	userId, _ := strconv.Atoi(ps.ByName("userId"))
	messageId, _ := strconv.Atoi(ps.ByName("messageId"))

	if !rt.checkAuth(w, r, userId) {
		return
	}

	var reqBody struct {
		Targets []int `json:"targets"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido")
		return
	}

	originalMsg, err := rt.db.GetMessage(messageId)
	if err != nil {
		rt.sendError(w, http.StatusNotFound, 404, "Messaggio originale non trovato")
		return
	}

	forwardTime := time.Now()

	for _, tid := range reqBody.Targets {
		_, _ = rt.db.CreateMessage(tid, userId, originalMsg.Text, originalMsg.PhotoUrl, forwardTime, 0, true)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(originalMsg)
}
