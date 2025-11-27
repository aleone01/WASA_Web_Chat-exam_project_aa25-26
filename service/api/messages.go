package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/aleone01/Web-Project-repo/service/api/reqcontext"
	"github.com/aleone01/Web-Project-repo/service/globaltime"
	"github.com/julienschmidt/httprouter"
)

// sendMessage invia un nuovo messaggio
func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Estrazione parametri da URL
	userId, _ := strconv.Atoi(ps.ByName("userId"))
	chatId, _ := strconv.Atoi(ps.ByName("chatId"))

	// Check Auth
	if !rt.checkAuth(w, r, userId) {
		return
	}

	// Parsing Multipart
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "Errore nel parsing del messaggio")
		return
	}

	text := r.FormValue("text")
	var photoUrl string

	file, fileHeader, err := r.FormFile("photo")
	if err == nil {
		defer file.Close()

		storagePath := "./images"
		// Ignoriamo l'errore di MkdirAll se la cartella esiste già
		_ = os.MkdirAll(storagePath, 0755)

		filename := fmt.Sprintf("msg_%d_%d%s", userId, globaltime.Now().Unix(), filepath.Ext(fileHeader.Filename))
		fullPath := filepath.Join(storagePath, filename)

		dst, err := os.Create(fullPath)
		if err != nil {
			rt.sendError(w, http.StatusInternalServerError, 500, "Errore salvataggio file")
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			rt.sendError(w, http.StatusInternalServerError, 500, "Errore scrittura file")
			return
		}

		photoUrl = "/images/" + filename
	}

	msg, err := rt.db.CreateMessage(chatId, userId, text, photoUrl)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore nel salvataggio del messaggio")
		return
	}

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

	for _, targetChatId := range reqBody.Targets {
		// Ignoriamo errori sui singoli invii per semplicità, ma logghiamo se possibile
		_, _ = rt.db.CreateMessage(targetChatId, userId, originalMsg.Text, originalMsg.PhotoUrl)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(originalMsg)
}
