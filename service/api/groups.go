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

// createGroup gestisce la creazione di un nuovo gruppo
func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Auth manuale semplificata (recupera user dal token)
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

	// parsing del body della richiesta
	var reqBody struct {
		Name    string `json:"name"`
		Photo   string `json:"photo"`
		Members []int  `json:"members"`
	}

	// creazione del gruppo nel database
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido")
		return
	}

	g, err := rt.db.CreateGroup(reqBody.Name, reqBody.Photo, reqBody.Members, userId)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore creazione")
		return
	}

	// risposta con i dati del gruppo creato
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(g)
}

// addToGroup gestisce l'aggiunta di membri a un gruppo
func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// estrazione groupId dai parametri della rotta
	groupId, _ := strconv.Atoi(ps.ByName("groupId"))

	// Auth manuale semplificata (recupera user dal token)
	token := r.Header.Get("Authorization")
	if len(token) > 7 {
		token = token[7:]
	} else if token == "" {
		token = r.Header.Get("sessionToken")
	}
	userId, _ := rt.db.CheckToken(token)

	// verifica se l'utente è membro del gruppo
	isMember, _ := rt.db.CheckGroupMembership(groupId, userId)
	if !isMember {
		rt.sendError(w, http.StatusForbidden, 403, "Non sei membro")
		return
	}

	// parsing del body della richiesta
	var reqBody struct {
		Members []int `json:"members"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido")
		return
	}

	// aggiunta dei membri al gruppo nel database
	g, _ := rt.db.AddGroupMembers(groupId, reqBody.Members)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(g)
}

// leaveGroup gestisce l'uscita di un utente da un gruppo
func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// estrazione groupId dai parametri della rotta
	groupId, _ := strconv.Atoi(ps.ByName("groupId"))
	token := r.Header.Get("Authorization")

	// Auth manuale semplificata (recupera user dal token)
	if len(token) > 7 {
		token = token[7:]
	} else if token == "" {
		token = r.Header.Get("sessionToken")
	}
	userId, _ := rt.db.CheckToken(token)

	// rimozione dell'utente dal gruppo nel database
	if err := rt.db.LeaveGroup(groupId, userId); err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore")
		return
	}

	// risposta di successo senza contenuto
	w.WriteHeader(http.StatusNoContent)
}

// setGroupName gestisce l'aggiornamento del nome del gruppo
func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// estrazione groupId dai parametri della rotta
	groupId, _ := strconv.Atoi(ps.ByName("groupId"))

	// Auth manuale semplificata (recupera user dal token)
	token := r.Header.Get("Authorization")
	if len(token) > 7 {
		token = token[7:]
	} else if token == "" {
		token = r.Header.Get("sessionToken")
	}
	userId, _ := rt.db.CheckToken(token)

	// verifica se l'utente è membro del gruppo
	isMember, _ := rt.db.CheckGroupMembership(groupId, userId)
	if !isMember {
		rt.sendError(w, http.StatusForbidden, 403, "Accesso negato")
		return
	}

	// parsing del body della richiesta
	var reqBody struct {
		Name string `json:"name"`
	}

	// aggiornamento del nome del gruppo nel database
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido")
		return
	}
	g, _ := rt.db.SetGroupName(groupId, reqBody.Name)

	// risposta con i dati aggiornati del gruppo
	w.Header().Set("Content-Type", "application/json")
	// CORREZIONE: Assegnato a _
	_ = json.NewEncoder(w).Encode(g)
}

// setGroupPhoto gestisce l'aggiornamento della foto del gruppo
func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// estrazione groupId dai parametri della rotta
	groupId, _ := strconv.Atoi(ps.ByName("groupId"))

	// Auth manuale semplificata (recupera user dal token)
	token := r.Header.Get("Authorization")
	if len(token) > 7 {
		token = token[7:]
	} else if token == "" {
		token = r.Header.Get("sessionToken")
	}
	userId, _ := rt.db.CheckToken(token)

	// verifica se l'utente è membro del gruppo
	isMember, _ := rt.db.CheckGroupMembership(groupId, userId)
	if !isMember {
		rt.sendError(w, http.StatusForbidden, 403, "Accesso negato")
		return
	}

	// parsing del form multipart per l'immagine
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "Errore form")
		return
	}

	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "File mancante")
		return
	}
	defer file.Close()

	// salvataggio del file immagine
	filename := fmt.Sprintf("g%d_%d%s", groupId, globaltime.Now().Unix(), filepath.Ext(fileHeader.Filename))

	// Creazione directory se non esiste
	if err := os.MkdirAll("./images", 0755); err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore filesystem")
		return
	}

	path := filepath.Join("./images", filename)

	dst, err := os.Create(path)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore salvataggio")
		return
	}
	defer dst.Close()

	// copia del contenuto del file caricato
	if _, err := io.Copy(dst, file); err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore scrittura file")
		return
	}

	// aggiornamento della foto del gruppo nel database
	g, _ := rt.db.SetGroupPhoto(groupId, "/images/"+filename)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(g)
}
