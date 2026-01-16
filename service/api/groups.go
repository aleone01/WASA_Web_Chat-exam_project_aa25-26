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
// Flusso operativo:
// 1. Estrazione e validazione manuale del token (Header o SessionToken).
// 2. Parsing del form multipart (Nome gruppo, Lista membri, Foto opzionale).
// 3. Risoluzione degli username dei membri in ID utente, verificando che esistano.
// 4. Creazione del gruppo nel DB.
func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Estrazione manuale del token per validazione (supporta Bearer o header custom)
	token := r.Header.Get("Authorization")
	if len(token) > 7 {
		token = token[7:] // Rimuove il prefisso "Bearer "
	} else if token == "" {
		token = r.Header.Get("sessionToken")
	}

	// Validazione token tramite DB
	userId, err := rt.db.CheckToken(token)
	if err != nil {
		rt.sendError(w, http.StatusUnauthorized, 401, "Login richiesto o token non valido")
		return
	}

	// Parsing del form (max 10MB per gestire eventuale foto gruppo)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "Errore nel parsing del form data")
		return
	}

	// Recupero dati testuali obbligatori
	groupName := r.FormValue("groupname")
	if groupName == "" {
		rt.sendError(w, http.StatusBadRequest, 400, "Nome del gruppo mancante")
		return
	}

	// Parsing della lista membri fornita come stringa JSON
	membersStr := r.FormValue("membersList")
	var memberUsernames []string
	if membersStr != "" {
		if err := json.Unmarshal([]byte(membersStr), &memberUsernames); err != nil {
			rt.sendError(w, http.StatusBadRequest, 400, "Formato membersList non valido (atteso JSON array di stringhe)")
			return
		}
	}

	// Risoluzione degli Username in ID numerici e validazione esistenza utenti
	var memberIds []int
	for _, username := range memberUsernames {
		u, err := rt.db.GetUserByUsername(username)
		if err != nil {
			// Fail-fast se anche solo un utente non esiste
			rt.sendError(w, http.StatusBadRequest, 400, "Utente indicato non esistente: "+username)
			return
		}
		// Evita di aggiungere il creatore due volte (viene aggiunto automaticamente dalla logica DB solitamente)
		if u.Id == userId {
			continue
		}
		memberIds = append(memberIds, u.Id)
	}

	// Gestione foto del gruppo (opzionale)
	var photoFile []byte
	file, _, err := r.FormFile("groupPhotoFile")
	if err == nil {
		defer file.Close()
		photoFile, err = io.ReadAll(file)
		if err != nil {
			rt.sendError(w, http.StatusInternalServerError, 500, "Errore durante la lettura della foto")
			return
		}
	}

	// Creazione su DB passando gli ID validati e i metadati
	g, err := rt.db.CreateGroup(groupName, photoFile, memberIds, userId)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore durante la creazione del gruppo")
		return
	}

	// Risposta successo
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(g)
}

// addToGroup permette di aggiungere un utente esistente ad un gruppo.
// Richiede che chi esegue l'azione sia già membro del gruppo.
func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	groupId, _ := strconv.Atoi(ps.ByName("groupId"))

	// Verifica autenticazione
	token := r.Header.Get("Authorization")
	if len(token) > 7 {
		token = token[7:]
	} else if token == "" {
		token = r.Header.Get("sessionToken")
	}
	userId, _ := rt.db.CheckToken(token)

	// Controllo autorizzativo: solo i membri possono aggiungere altri utenti
	isMember, _ := rt.db.CheckGroupMembership(groupId, userId)
	if !isMember {
		rt.sendError(w, http.StatusForbidden, 403, "Non sei membro di questo gruppo, azione negata")
		return
	}

	// Parsing body per ottenere l'username da aggiungere
	var reqBody struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido")
		return
	}

	// Risoluzione utente target
	targetUser, err := rt.db.GetUserByUsername(reqBody.Username)
	if err != nil {
		rt.sendError(w, http.StatusNotFound, 404, "Utente da aggiungere non trovato")
		return
	}

	// Aggiornamento DB
	g, _ := rt.db.AddGroupMembers(groupId, []int{targetUser.Id})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(g)
}

// leaveGroup permette a un utente di abbandonare un gruppo.
func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	groupId, _ := strconv.Atoi(ps.ByName("groupId"))
	token := r.Header.Get("Authorization")

	if len(token) > 7 {
		token = token[7:]
	} else if token == "" {
		token = r.Header.Get("sessionToken")
	}
	userId, _ := rt.db.CheckToken(token)

	// Esecuzione logica di uscita
	if err := rt.db.LeaveGroup(groupId, userId); err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore durante l'uscita dal gruppo")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// setGroupName permette di rinominare un gruppo.
// Richiede appartenenza al gruppo.
func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	groupId, _ := strconv.Atoi(ps.ByName("groupId"))

	// Auth Check
	token := r.Header.Get("Authorization")
	if len(token) > 7 {
		token = token[7:]
	} else if token == "" {
		token = r.Header.Get("sessionToken")
	}
	userId, _ := rt.db.CheckToken(token)

	// Membership Check
	isMember, _ := rt.db.CheckGroupMembership(groupId, userId)
	if !isMember {
		rt.sendError(w, http.StatusForbidden, 403, "Accesso negato: non sei membro")
		return
	}

	// Parsing del nuovo nome
	var reqBody struct {
		Name string `json:"groupname"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido")
		return
	}

	// Aggiornamento DB
	g, _ := rt.db.SetGroupName(groupId, reqBody.Name)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(g)
}

// setGroupPhoto aggiorna l'immagine del gruppo.
// Richiede multipart form e appartenenza al gruppo.
func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	groupId, _ := strconv.Atoi(ps.ByName("groupId"))

	// Auth Check
	token := r.Header.Get("Authorization")
	if len(token) > 7 {
		token = token[7:]
	} else if token == "" {
		token = r.Header.Get("sessionToken")
	}
	userId, _ := rt.db.CheckToken(token)

	// Membership Check
	isMember, _ := rt.db.CheckGroupMembership(groupId, userId)
	if !isMember {
		rt.sendError(w, http.StatusForbidden, 403, "Accesso negato: non sei membro")
		return
	}

	// Parsing form
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "Errore form")
		return
	}

	// Lettura file
	file, _, err := r.FormFile("groupPhotoFile")
	if err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "File mancante")
		return
	}
	defer file.Close()

	// Lettura byte file
	photoFile, err := io.ReadAll(file)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore lettura file")
		return
	}

	// Aggiornamento DB
	g, _ := rt.db.SetGroupPhoto(groupId, photoFile)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(g)
}

// getGroupMembers restituisce la lista degli utenti appartenenti al gruppo.
// Solo i membri possono vedere chi altro c'è nel gruppo.
func (rt *_router) getGroupMembers(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	groupId, _ := strconv.Atoi(ps.ByName("groupId"))

	// Auth Check
	token := r.Header.Get("Authorization")
	if len(token) > 7 {
		token = token[7:]
	} else if token == "" {
		token = r.Header.Get("sessionToken")
	}
	userId, _ := rt.db.CheckToken(token)

	// Controllo membership: privacy del gruppo
	isMember, _ := rt.db.CheckGroupMembership(groupId, userId)
	if !isMember {
		rt.sendError(w, http.StatusForbidden, 403, "Non sei membro del gruppo, impossibile visualizzare i partecipanti")
		return
	}

	// Recupero lista dal DB
	members, err := rt.db.GetGroupMembers(groupId)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore DB nel recupero membri")
		return
	}

	// Wrapping della risposta in un oggetto JSON
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		Members interface{} `json:"membersList"`
	}{Members: members})
}
