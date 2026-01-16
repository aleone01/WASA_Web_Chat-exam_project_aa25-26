package api

import (
	"encoding/json"
	"net/http"

	"github.com/aleone01/Web-Project-repo/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// doLogin gestisce il processo di autenticazione o registrazione di un utente (pattern "Upsert" logico).
// 1. Riceve un username nel corpo della richiesta.
// 2. Valida la lunghezza dell'username.
// 3. Interroga il database:
//   - Se l'utente esiste, restituisce l'identificativo esistente (Login).
//   - Se l'utente non esiste, lo crea e restituisce il nuovo identificativo (Registrazione).
//
// Restituisce 200 OK se l'utente esisteva, 201 Created se è stato appena creato.
func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Struttura per il parsing del JSON in arrivo
	var reqBody struct {
		Username *string `json:"username"`
	}

	// Parsing del body
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido o malformato")
		return
	}

	// Validazione delle regole di business sullo username
	if reqBody.Username == nil || len(*reqBody.Username) < 3 || len(*reqBody.Username) > 16 {
		rt.sendError(w, http.StatusBadRequest, 400, "Username non valido (deve essere tra 3 e 16 caratteri)")
		return
	}

	// Logging informativo dell'operazione
	ctx.Logger.Infof("Tentativo di login/registrazione per utente: %s", *reqBody.Username)

	// Chiamata al layer database che gestisce la logica "Get or Create"
	id, created, err := rt.db.UserLogin(*reqBody.Username)
	if err != nil {
		ctx.Logger.WithError(err).Error("Errore durante l'interazione con il DB per il Login")
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore interno del server")
		return
	}

	// Preparazione della risposta JSON
	w.Header().Set("Content-Type", "application/json")

	// Impostazione del codice di stato HTTP corretto in base all'esito
	if created {
		w.WriteHeader(http.StatusCreated) // 201 Created
	} else {
		w.WriteHeader(http.StatusOK) // 200 OK
	}

	// Invio dell'identificativo utente (session token/id)
	_ = json.NewEncoder(w).Encode(struct {
		Identifier int `json:"identifier"`
	}{Identifier: id})
}
