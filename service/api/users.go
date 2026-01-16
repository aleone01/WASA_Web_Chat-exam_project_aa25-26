package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/aleone01/Web-Project-repo/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// setMyUserName permette all'utente autenticato di modificare il proprio nome visualizzato (username).
// La funzione prevede i seguenti passaggi:
// 1. Verifica dell'autenticazione tramite ID utente.
// 2. Parsing del corpo della richiesta JSON.
// 3. Validazione semantica del nuovo username (lunghezza tra 3 e 16 caratteri).
// 4. Aggiornamento sul database.
// Restituisce l'oggetto utente aggiornato in caso di successo.
func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Estrazione dell'ID utente dai parametri del percorso
	userId, _ := strconv.Atoi(ps.ByName("userId"))

	// Verifica se l'utente che fa la richiesta è autorizzato a modificare questo profilo
	if !rt.checkAuth(w, r, userId) {
		return
	}

	// Struttura temporanea per il decoding del JSON in ingresso
	var reqBody struct {
		Username *string `json:"username"`
	}

	// Tentativo di decodifica del body JSON
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido: formato errato")
		return
	}

	// Validazione input: il campo deve essere presente e rispettare i limiti di lunghezza
	if reqBody.Username == nil || len(*reqBody.Username) < 3 || len(*reqBody.Username) > 16 {
		rt.sendError(w, http.StatusBadRequest, 400, "Username non valido: deve essere compreso tra 3 e 16 caratteri")
		return
	}

	// Chiamata al database per aggiornare lo username
	user, err := rt.db.SetMyUsername(userId, *reqBody.Username)
	if err != nil {
		// Log dell'errore e risposta generica al client
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore durante l'aggiornamento dello username")
		return
	}

	// Risposta positiva con l'oggetto User aggiornato
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}

// setMyPhoto gestisce l'aggiornamento dell'immagine del profilo dell'utente.
// Utilizza una richiesta multipart/form-data.
// La funzione si occupa di leggere lo stream del file, convertirlo in array di byte e salvarlo nel DB.
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	userId, _ := strconv.Atoi(ps.ByName("userId"))

	// Verifica autorizzazione: solo il proprietario può cambiare la foto
	if !rt.checkAuth(w, r, userId) {
		return
	}

	// Parsing del form multipart.
	// Viene impostato un limite di memoria di 10MB (10 << 20). File più grandi potrebbero essere rifiutati.
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "Errore nel parsing del form data o file troppo grande (max 10MB)")
		return
	}

	// Recupero del file con chiave "photo" dal form
	file, _, err := r.FormFile("photo")
	if err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "Campo file 'photo' mancante nella richiesta")
		return
	}
	defer file.Close() // Assicura la chiusura del file descriptor alla fine della funzione

	// Lettura completa del contenuto del file in un slice di byte ([]byte)
	photoFile, err := io.ReadAll(file)
	if err != nil {
		ctx.Logger.WithError(err).Error("Errore durante la lettura dello stream dell'immagine")
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore interno durante la lettura del file")
		return
	}

	// Aggiornamento del record utente nel database con il nuovo blob dell'immagine
	user, err := rt.db.SetProfilePhoto(userId, photoFile)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore durante il salvataggio della foto nel database")
		return
	}

	// Invio della risposta con l'utente aggiornato
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}
