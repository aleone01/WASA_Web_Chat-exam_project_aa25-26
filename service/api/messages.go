package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/aleone01/Web-Project-repo/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// sendMessage gestisce l'invio di un nuovo messaggio.
func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	userId, _ := strconv.Atoi(ps.ByName("userId"))
	chatId, _ := strconv.Atoi(ps.ByName("chatId"))

	if !rt.checkAuth(w, r, userId) {
		return
	}

	// Parsing multipart (max 10 MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "Errore dati form (usa multipart/form-data)")
		return
	}

	// Recupero campi testuali dal form
	text := r.FormValue("text")
	replyTo, _ := strconv.Atoi(r.FormValue("replyTo"))
	isForward, _ := strconv.ParseBool(r.FormValue("isForward"))

	// Gestione Foto (opzionale)
	var photoFile []byte

	// Proviamo a recuperare il file
	file, _, err := r.FormFile("file")
	if err == nil {
		defer file.Close()
		photoFile, err = io.ReadAll(file)
		if err != nil {
			rt.sendError(w, http.StatusInternalServerError, 500, "Errore lettura file")
			return
		}
	} else if !errors.Is(err, http.ErrMissingFile) {

		rt.sendError(w, http.StatusInternalServerError, 500, "Errore form file")
		return
	}

	// Se è un messaggio inoltrato, recuperiamo il messaggio originale
	if isForward && replyTo > 0 {
		originalMsg, err := rt.db.GetMessage(replyTo)
		if err == nil {
			if text == "" {
				text = originalMsg.Text
			}
			if len(photoFile) == 0 {
				photoFile = originalMsg.PhotoFile
			}
		} else {
			rt.sendError(w, http.StatusNotFound, 404, "Messaggio originale da inoltrare non trovato")
			return
		}
	}

	if text == "" && len(photoFile) == 0 {
		rt.sendError(w, http.StatusBadRequest, 400, "Messaggio vuoto")
		return
	}

	sentAt := time.Now()

	msg, err := rt.db.CreateMessage(chatId, userId, text, photoFile, sentAt, replyTo, isForward)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore nel salvataggio")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(msg)
}

// commentMessage aggiunge reazione
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

// uncommentMessage rimuove reazione
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

// deleteMessage cancella messaggio
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

// forwardMessage inoltra messaggio
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
		_, _ = rt.db.CreateMessage(tid, userId, originalMsg.Text, originalMsg.PhotoFile, forwardTime, 0, true)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(originalMsg)
}
