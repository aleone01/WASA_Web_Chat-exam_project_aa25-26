package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/aleone01/Web-Project-repo/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// createGroup permette a un utente autenticato di creare un nuovo gruppo di conversazione.
// La funzione estrae il token di autenticazione per identificare il creatore, analizza il corpo della richiesta JSON
// per ottenere il nome del gruppo, la foto e la lista iniziale dei membri, e infine invoca il metodo del database
// per persistere il nuovo gruppo. Restituisce l'oggetto gruppo creato o un errore appropriato.
func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	token := r.Header.Get("Authorization")
	if len(token) > 7 {
		token = token[7:]
	} else if token == "" {
		token = r.Header.Get("sessionToken")
	}

	userId, err := rt.db.CheckToken(token)
	if err != nil {
		rt.sendError(w, http.StatusUnauthorized, 401, "Login richiesto")
		return
	}

	var reqBody struct {
		Name    string `json:"groupname"`
		Photo   string `json:"groupPhoto"`
		Members []int  `json:"membersList"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido")
		return
	}

	g, err := rt.db.CreateGroup(reqBody.Name, reqBody.Photo, reqBody.Members, userId)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore creazione")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(g)
}

// addToGroup consente l'aggiunta di nuovi membri a un gruppo esistente.
// Verifica preliminarmente che l'utente che effettua la richiesta sia già un membro del gruppo (controllo dei permessi).
// Se autorizzato, decodifica la lista dei nuovi membri dal corpo della richiesta e aggiorna il database.
func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	groupId, _ := strconv.Atoi(ps.ByName("groupId"))

	token := r.Header.Get("Authorization")
	if len(token) > 7 {
		token = token[7:]
	} else if token == "" {
		token = r.Header.Get("sessionToken")
	}
	userId, _ := rt.db.CheckToken(token)

	isMember, _ := rt.db.CheckGroupMembership(groupId, userId)
	if !isMember {
		rt.sendError(w, http.StatusForbidden, 403, "Non sei membro")
		return
	}

	var reqBody struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido")
		return
	}

	targetUser, err := rt.db.GetUserByUsername(reqBody.Username)
	if err != nil {
		rt.sendError(w, http.StatusNotFound, 404, "Utente non trovato")
		return
	}

	g, _ := rt.db.AddGroupMembers(groupId, []int{targetUser.Id})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(g)
}

// leaveGroup gestisce la richiesta di un utente di abbandonare un gruppo specifico.
// Identifica l'utente tramite il token di autenticazione e invoca la logica del database per rimuovere
// l'associazione tra l'utente e il gruppo indicato nell'URL.
func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	groupId, _ := strconv.Atoi(ps.ByName("groupId"))
	token := r.Header.Get("Authorization")

	if len(token) > 7 {
		token = token[7:]
	} else if token == "" {
		token = r.Header.Get("sessionToken")
	}
	userId, _ := rt.db.CheckToken(token)

	if err := rt.db.LeaveGroup(groupId, userId); err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// setGroupName permette di modificare il nome di un gruppo esistente.
// Esegue un controllo di appartenenza per assicurarsi che solo i membri del gruppo possano modificarne le proprietà.
// In caso di successo, aggiorna il nome nel database e restituisce l'oggetto gruppo aggiornato.
func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	groupId, _ := strconv.Atoi(ps.ByName("groupId"))

	token := r.Header.Get("Authorization")
	if len(token) > 7 {
		token = token[7:]
	} else if token == "" {
		token = r.Header.Get("sessionToken")
	}
	userId, _ := rt.db.CheckToken(token)

	isMember, _ := rt.db.CheckGroupMembership(groupId, userId)
	if !isMember {
		rt.sendError(w, http.StatusForbidden, 403, "Accesso negato")
		return
	}

	var reqBody struct {
		Name string `json:"groupname"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido")
		return
	}
	g, _ := rt.db.SetGroupName(groupId, reqBody.Name)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(g)
}

// setGroupPhoto consente di aggiornare l'immagine (o il riferimento all'immagine) di un gruppo.
// Similmente alle altre operazioni di modifica, verifica che il richiedente sia membro del gruppo prima di
// procedere con l'aggiornamento nel database.
func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	groupId, _ := strconv.Atoi(ps.ByName("groupId"))

	token := r.Header.Get("Authorization")
	if len(token) > 7 {
		token = token[7:]
	} else if token == "" {
		token = r.Header.Get("sessionToken")
	}
	userId, _ := rt.db.CheckToken(token)

	isMember, _ := rt.db.CheckGroupMembership(groupId, userId)
	if !isMember {
		rt.sendError(w, http.StatusForbidden, 403, "Accesso negato")
		return
	}

	var reqBody struct {
		Photo string `json:"groupPhoto"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido")
		return
	}

	g, _ := rt.db.SetGroupPhoto(groupId, reqBody.Photo)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(g)

}

func (rt *_router) getGroupMembers(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	groupId, _ := strconv.Atoi(ps.ByName("groupId"))

	token := r.Header.Get("Authorization")
	if len(token) > 7 {
		token = token[7:]
	} else if token == "" {
		token = r.Header.Get("sessionToken")
	}
	userId, _ := rt.db.CheckToken(token)

	// Controllo membership
	isMember, _ := rt.db.CheckGroupMembership(groupId, userId)
	if !isMember {
		rt.sendError(w, http.StatusForbidden, 403, "Non sei membro del gruppo")
		return
	}

	members, err := rt.db.GetGroupMembers(groupId)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore DB")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		Members interface{} `json:"membersList"`
	}{Members: members})
}
