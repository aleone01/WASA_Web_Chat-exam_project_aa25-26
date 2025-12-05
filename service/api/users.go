package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/aleone01/Web-Project-repo/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// setMyUserName gestisce l'aggiornamento del nome utente
func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userId, _ := strconv.Atoi(ps.ByName("userId"))

	// Check autenticazione
	if !rt.checkAuth(w, r, userId) {
		return
	}

	// Parsing JSON
	var reqBody struct {
		Username *string `json:"username"`
	}

	// Decodifica JSON
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido")
		return
	}

	// Validazione username
	if reqBody.Username == nil || len(*reqBody.Username) < 3 || len(*reqBody.Username) > 16 {
		rt.sendError(w, http.StatusBadRequest, 400, "Username non valido")
		return
	}

	// Aggiornamento database
	user, err := rt.db.SetMyUsername(userId, *reqBody.Username)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore aggiornamento username")
		return
	}

	// Risposta con il nuovo utente
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)

}

// setMyPhoto gestisce l'aggiornamento della foto del profilo utente (URL)
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Estrae userId dai parametri
	userId, _ := strconv.Atoi(ps.ByName("userId"))

	// Check autenticazione
	if !rt.checkAuth(w, r, userId) {
		return
	}

	// Parsing JSON
	var reqBody struct {
		Photo string `json:"photo"` // URL della nuova foto
	}

	// Decodifica JSON
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido")
		return
	}

	// Validazione URL minima
	if len(reqBody.Photo) == 0 {
		rt.sendError(w, http.StatusBadRequest, 400, "URL foto mancante")
		return
	}

	// Aggiornamento database
	user, err := rt.db.SetProfilePhoto(userId, reqBody.Photo)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore DB")
		return
	}

	// Risposta con il nuovo utente
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)

}
