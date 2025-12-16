package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/aleone01/Web-Project-repo/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// setMyUserName permette all'utente autenticato di modificare il proprio nome visualizzato (username).
// La funzione valida il nuovo nome (lunghezza minima e massima) e, se valido, aggiorna il record utente nel database.
func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userId, _ := strconv.Atoi(ps.ByName("userId"))

	if !rt.checkAuth(w, r, userId) {
		return
	}

	var reqBody struct {
		Username *string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido")
		return
	}

	if reqBody.Username == nil || len(*reqBody.Username) < 3 || len(*reqBody.Username) > 16 {
		rt.sendError(w, http.StatusBadRequest, 400, "Username non valido")
		return
	}

	user, err := rt.db.SetMyUsername(userId, *reqBody.Username)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore aggiornamento username")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)

}

// setMyPhoto gestisce l'aggiornamento dell'immagine del profilo dell'utente.
// Accetta un URL nel corpo della richiesta, verifica che non sia vuoto e aggiorna il riferimento fotografico
// associato all'utente nel database.
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	userId, _ := strconv.Atoi(ps.ByName("userId"))

	if !rt.checkAuth(w, r, userId) {
		return
	}

	var reqBody struct {
		Photo string `json:"photo"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido")
		return
	}

	if len(reqBody.Photo) == 0 {
		rt.sendError(w, http.StatusBadRequest, 400, "URL foto mancante")
		return
	}

	user, err := rt.db.SetProfilePhoto(userId, reqBody.Photo)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore DB")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)

}
