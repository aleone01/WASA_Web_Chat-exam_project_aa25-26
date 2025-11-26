package api

import 
(
	"github.com/aleone01/Web-Project-repo/service/globaltime"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// SendMessage invia un nuovo messaggio in una conversazione (con foto opzionale, POST /users/{userId}/chats/{chatId}/messages)
func (rt *Router) SendMessage(w http.ResponseWriter, r *http.Request, userId int, chatId int) 
{
	// Check Auth
	if err := rt.checkAuth(w, r, userId); err != nil 
	{
		return
	}

	// parsing del form multipart (Max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil 
	{
		rt.sendError(w, http.StatusBadRequest, 400, "Errore nel parsing del messaggio")
		return
	}

	// recupero del testo
	text := r.FormValue("text")

	// recupero della foto (opzionale)
	var photoUrl string
	file, fileHeader, err := r.FormFile("photo")
	
	// se è stata inviata una foto 
	if err == nil 
	{
		defer file.Close()

		// creazione cartella storage
		storagePath := "./images"
		if err := os.MkdirAll(storagePath, os.ModePerm); err != nil 
		{
			rt.sendError(w, http.StatusInternalServerError, 500, "Errore filesystem server")
			return
		}

		// generazione nome univoco con globaltime
		filename := fmt.Sprintf("msg_%d_%d%s", userId, globaltime.Now().Unix(), filepath.Ext(fileHeader.Filename))
		fullPath := filepath.Join(storagePath, filename)

		// creazione file su disco
		dst, err := os.Create(fullPath)
		if err != nil 
		{
			rt.sendError(w, http.StatusInternalServerError, 500, "Errore salvataggio file")
			return
		}
		defer dst.Close()

		// scrittura del file
		if _, err := io.Copy(dst, file); err != nil 
		{
			rt.sendError(w, http.StatusInternalServerError, 500, "Errore scrittura file")
			return
		}

		// URL relativo per il DB 
		photoUrl = "/images/" + filename
	}

	// chiamata al database per creare il messaggio
	message, err := rt.db.CreateMessage(chatId, userId, text, photoUrl)
	if err != nil 
	{
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore nel salvataggio del messaggio")
		return
	}

	// risposta JSON con il messaggio creato
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	json.NewEncoder(w).Encode(message)
}

// CommentMessage aggiunge una reazione a un messaggio (POST .../messages/{messageId}/reactions)
func (rt *Router) CommentMessage(w http.ResponseWriter, r *http.Request, userId int, chatId int, messageId int) 
{
	// Check Auth
	if err := rt.checkAuth(w, r, userId); err != nil 
	{
		return
	}

	// decodifica body JSON 
	var reqBody struct 
	{
		Emoticon string `json:"emoticon"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil 
	{
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido")
		return
	}

	// Chiamata al DB per aggiungere la reazione
	err := rt.db.AddReaction(messageId, userId, reqBody.Emoticon)
	if err != nil 
	{
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore nell'aggiunta della reazione")
		return
	}

	// risposta con successo 
	w.WriteHeader(http.StatusCreated)
}

// UncommentMessage rimuove una reazione (DELETE .../messages/{messageId}/reactions)
func (rt *Router) UncommentMessage(w http.ResponseWriter, r *http.Request, userId int, chatId int, messageId int) 
{
	// Check Auth
	if err := rt.checkAuth(w, r, userId); err != nil 
	{
		return
	}

	// chiamata al DB per rimuovere la reazione
	err := rt.db.RemoveReaction(messageId, userId)
	if err != nil 
	{
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore nella rimozione reazione")
		return
	}

	// risposta con successo 
	w.WriteHeader(http.StatusNoContent)
}

// DeleteMessage elimina un messaggio (DELETE .../messages/{messageId})
func (rt *Router) DeleteMessage(w http.ResponseWriter, r *http.Request, userId int, chatId int, messageId int) 
{

	// Check Auth
	if err := rt.checkAuth(w, r, userId); err != nil 
	{
		return
	}

	// chiamata al DB per eliminare il messaggio
	err := rt.db.DeleteMessage(messageId, userId)
	if err != nil 
	{
		// se il messaggio non esiste o non appartiene all'utente
		if err.Error() == "message not found or unauthorized" 
		{
			rt.sendError(w, http.StatusForbidden, 403, "Non autorizzato a cancellare questo messaggio")
			return
		}
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore nella cancellazione")
		return
	}

	// risposta con successo
	w.WriteHeader(http.StatusNoContent) 
}

// ForwardMessage inoltra un messaggio ad altre chat (POST .../messages/{messageId}/forward)
func (rt *Router) ForwardMessage(w http.ResponseWriter, r *http.Request, userId int, chatId int, messageId int) 
{
	// Check Auth
	if err := rt.checkAuth(w, r, userId); err != nil 
	{
		return
	}

	// decodifica body JSON con lista di chat targettate
	var reqBody struct 
	{
		Targets []int `json:"targets"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil 
	{
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido")
		return
	}

	// recupera il messaggio originale
	originalMsg, err := rt.db.GetMessage(messageId)
	if err != nil 
	{
		rt.sendError(w, http.StatusNotFound, 404, "Messaggio originale non trovato")
		return
	}

	// inoltra il messaggio a ciascuna chat targettata
	for _, targetChatId := range reqBody.Targets 
	{
		// crea una copia del messaggio nella nuova chat
		_, err := rt.db.CreateMessage(targetChatId, userId, originalMsg.Text, originalMsg.PhotoUrl)
		if err != nil 
		{
			rt.sendError(w, http.StatusInternalServerError, 500, "Errore durante l'inoltro")
			return
		}
	}

	// risposta di successo
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

}