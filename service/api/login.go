package api

import (
	"encoding/json"
	"net/http"

	"github.com/aleone01/Web-Project-repo/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// doLogin gestisce il processo di autenticazione o registrazione di un utente.
// La funzione accetta un nome utente nel corpo della richiesta, ne valida la lunghezza, e tenta di effettuare il login tramite il database.
// Se l'utente non esiste, viene creato (registrazione implicita). Restituisce l'identificativo univoco dell'utente
// e un codice di stato appropriato (200 OK per login esistente, 201 Created per nuovo utente).
func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var reqBody struct {
		Username *string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido")
		return
	}

	if reqBody.Username == nil || len(*reqBody.Username) < 3 || len(*reqBody.Username) > 16 {
		rt.sendError(w, http.StatusBadRequest, 400, "Username non valido (3-16 caratteri)")
		return
	}

	ctx.Logger.Infof("Tentativo di login per: %s", *reqBody.Username)

	id, created, err := rt.db.UserLogin(*reqBody.Username)
	if err != nil {
		ctx.Logger.WithError(err).Error("Errore DB Login")
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore interno")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if created {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	_ = json.NewEncoder(w).Encode(struct {
		Identifier int `json:"identifier"`
	}{Identifier: id})
}
