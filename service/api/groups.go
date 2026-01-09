package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/aleone01/Web-Project-repo/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// createGroup permette a un utente autenticato di creare un nuovo gruppo.
// Accetta una lista di username in "membersList". Verifica che tutti esistano prima di creare il gruppo.
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

	// Parsing del form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "Errore form data")
		return
	}

	// Recupero dati testuali
	groupName := r.FormValue("groupname")
	if groupName == "" {
		rt.sendError(w, http.StatusBadRequest, 400, "Nome gruppo mancante")
		return
	}

	// Parsing della lista membri
	membersStr := r.FormValue("membersList")
	var memberUsernames []string
	if membersStr != "" {
		if err := json.Unmarshal([]byte(membersStr), &memberUsernames); err != nil {
			rt.sendError(w, http.StatusBadRequest, 400, "Formato membersList non valido (atteso JSON array di stringhe)")
			return
		}
	}

	// Risoluzione degli Username in ID e validazione esistenza
	var memberIds []int
	for _, username := range memberUsernames {

		u, err := rt.db.GetUserByUsername(username)
		if err != nil {
			rt.sendError(w, http.StatusBadRequest, 400, "Utente non esistente: "+username)
			return
		}
		if u.Id == userId {
			continue
		}
		memberIds = append(memberIds, u.Id)
	}

	// Gestione foto opzionale
	var photoFile []byte
	file, _, err := r.FormFile("groupPhotoFile")
	if err == nil {
		defer file.Close()
		photoFile, err = io.ReadAll(file)
		if err != nil {
			rt.sendError(w, http.StatusInternalServerError, 500, "Errore lettura foto")
			return
		}
	}

	// Creazione su DB passando gli ID validati
	g, err := rt.db.CreateGroup(groupName, photoFile, memberIds, userId)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore creazione gruppo")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(g)
}

// addToGroup
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

// leaveGroup
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

// setGroupName
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

// setGroupPhoto
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

	// Parsing form
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "Errore form")
		return
	}

	file, _, err := r.FormFile("groupPhotoFile")
	if err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "File mancante")
		return
	}
	defer file.Close()

	// Leggi byte
	photoFile, err := io.ReadAll(file)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore lettura file")
		return
	}

	g, _ := rt.db.SetGroupPhoto(groupId, photoFile)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(g)
}

// getGroupMembers
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
