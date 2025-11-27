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

// setMyUserName gestisce l'aggiornamento del nome utente
func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// estrazione userId dai parametri della rotta
	userId, _ := strconv.Atoi(ps.ByName("userId"))

	// verifica autenticazione
	if !rt.checkAuth(w, r, userId) {
		return
	}

	// parsing del body della richiesta
	var reqBody struct {
		Username *string `json:"username"`
	}
	// in caso di errore nel parsing del JSON
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido")
		return
	}

	// validazione username
	if reqBody.Username == nil || len(*reqBody.Username) < 3 || len(*reqBody.Username) > 16 {
		rt.sendError(w, http.StatusBadRequest, 400, "Username non valido")
		return
	}

	// aggiornamento username nel database
	user, err := rt.db.SetMyUsername(userId, *reqBody.Username)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore aggiornamento username")
		return
	}

	// risposta con i dati aggiornati dell'utente
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}

// setMyPhoto gestisce l'aggiornamento della foto del profilo utente
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// estrazione userId dai parametri della rotta
	userId, _ := strconv.Atoi(ps.ByName("userId"))

	// verifica autenticazione
	if !rt.checkAuth(w, r, userId) {
		return
	}

	// parsing del form multipart
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "Errore form dati")
		return
	}

	// estrazione del file
	file, fileHeader, err := r.FormFile("photo")
	if err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "File mancante")
		return
	}
	defer file.Close()

	// salvataggio con globaltime e generazione del percorso
	storagePath := "./images"
	// Controllo errore MkdirAll
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore filesystem")
		return
	}

	// creazione del nome file unico
	filename := fmt.Sprintf("u%d_%d%s", userId, globaltime.Now().Unix(), filepath.Ext(fileHeader.Filename))
	fullPath := filepath.Join(storagePath, filename)

	// creazione del file sul disco
	dst, err := os.Create(fullPath)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore disco")
		return
	}
	defer dst.Close()

	// copia del contenuto nel file di destinazione
	// Controllo errore Copy
	if _, err := io.Copy(dst, file); err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore scrittura file")
		return
	}

	// aggiornamento del percorso della foto nel database
	user, err := rt.db.SetProfilePhoto(userId, "/images/"+filename)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore DB")
		return
	}

	// risposta con i dati aggiornati dell'utente
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}
