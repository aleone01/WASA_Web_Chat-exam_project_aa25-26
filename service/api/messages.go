package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/aleone01/Web-Project-repo/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// sendMessage invia un nuovo messaggio
func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// estrazione userId e chatId dai parametri della rotta
	userId, _ := strconv.Atoi(ps.ByName("userId"))
	chatId, _ := strconv.Atoi(ps.ByName("chatId"))

	// Check autenticazione
	if !rt.checkAuth(w, r, userId) {
		return
	}

	// Parsing JSON
	var reqBody struct {
		Text  string `json:"text"`
		Photo string `json:"photo"` // URL opzionale
	}

	// Decodifica JSON
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido")
		return
	}

	// Imposta il timestamp di invio
	sentAt := time.Now()

	// Creazione del messaggio nel database
	msg, err := rt.db.CreateMessage(chatId, userId, reqBody.Text, reqBody.Photo, sentAt)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore nel salvataggio")
		return
	}

	// Risposta con il messaggio creato
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(msg)
}

// commentMessage aggiunge una reazione
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

// uncommentMessage rimuove una reazione
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

// deleteMessage elimina un messaggio
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

// forwardMessage inoltra un messaggio
func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// estrazione userId e messageId dai parametri della rotta
	userId, _ := strconv.Atoi(ps.ByName("userId"))
	messageId, _ := strconv.Atoi(ps.ByName("messageId"))

	// Check autenticazione
	if !rt.checkAuth(w, r, userId) {
		return
	}

	// Parsing JSON
	var reqBody struct {
		Targets []int `json:"targets"`
	}

	// Decodifica JSON
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido")
		return
	}

	// Recupera il messaggio originale
	originalMsg, err := rt.db.GetMessage(messageId)
	if err != nil {
		rt.sendError(w, http.StatusNotFound, 404, "Messaggio originale non trovato")
		return
	}

	// Imposta il timestamp di inoltro
	forwardTime := time.Now()

	// Inoltra il messaggio a ciascun target
	for _, tid := range reqBody.Targets {
		_, _ = rt.db.CreateMessage(tid, userId, originalMsg.Text, originalMsg.PhotoUrl, forwardTime)
	}

	// Risposta con il messaggio originale inoltrato
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(originalMsg)

}
