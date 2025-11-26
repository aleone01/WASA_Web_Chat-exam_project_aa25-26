package api

import 
(
	"github.com/aleone01/Web-Project-repo/service/globaltime"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// CreateGroup crea un nuovo gruppo (POST /groups)
func (rt *Router) CreateGroup(w http.ResponseWriter, r *http.Request, userId int) 
{
	if err := rt.checkAuth(w, r, userId); err != nil 
	{
		return
	}

	// Parsing JSON body 
	var reqBody struct 
	{
		Name    string `json:"name"`
		Photo   string `json:"photo"` // YAML dice string binary, qui assumiamo URL o base64 string
		Members []int  `json:"members"`
	}

	// Decodifica JSON 
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil 
	{
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido")
		return
	}

	// controllo campi obbligatori
	if len(reqBody.Name) == 0 || len(reqBody.Members) == 0 
	{
		rt.sendError(w, http.StatusBadRequest, 400, "Nome e Membri richiesti")
		return
	}

	// chiamata al database per creare il gruppo
	group, err := rt.db.CreateGroup(reqBody.Name, reqBody.Photo, reqBody.Members, userId)
	if err != nil 
	{
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore creazione gruppo")
		return
	}

	// risposta JSON con il gruppo creato e messsaggio di successo
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(group)

}

// AddToGroup aggiunge membri a un gruppo esistente (POST /groups/{groupId}/members)
func (rt *Router) AddToGroup(w http.ResponseWriter, r *http.Request, userId int, groupId int) 
{

	// Check Auth
	if err := rt.checkAuth(w, r, userId); err != nil { return }

	// verifica che l'utente sia membro del gruppo
	isMember, _ := rt.db.CheckGroupMembership(groupId, userId)
	if !isMember 
	{
		rt.sendError(w, http.StatusForbidden, 403, "Non sei membro del gruppo")
		return
	}

	// parsing body JSON
	var reqBody struct 
	{
		Members []int `json:"members"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil 
	{
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido")
		return
	}

	// chiamata al database per aggiungere i membri
	updatedGroup, err := rt.db.AddGroupMembers(groupId, reqBody.Members)
	if err != nil 
	{
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore aggiunta membri")
		return
	}

	// risposta con il gruppo aggiornato e messaggio di successo
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedGroup)
}

// LeaveGroup abbandona il gruppo (DELETE /groups/{groupId}/leave)
func (rt *Router) LeaveGroup(w http.ResponseWriter, r *http.Request, userId int, groupId int) 
{

	// Check Auth
	if err := rt.checkAuth(w, r, userId); err != nil { return }

	// chiamata al database per lasciare il gruppo
	err := rt.db.LeaveGroup(groupId, userId)
	if err != nil 
	{
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore durante l'uscita (o non eri membro)")
		return
	}

	// risposta con successo
	w.WriteHeader(http.StatusNoContent)
}

// SetGroupName cambia il nome del gruppo (PUT /groups/{groupId}/name)
func (rt *Router) SetGroupName(w http.ResponseWriter, r *http.Request, userId int, groupId int) 
{
	// Check Auth
	if err := rt.checkAuth(w, r, userId); err != nil { return }

	// verifica che l'utente sia membro del gruppo
	isMember, _ := rt.db.CheckGroupMembership(groupId, userId)
	if !isMember 
	{
		rt.sendError(w, http.StatusForbidden, 403, "Non sei membro del gruppo")
		return
	}

	// parsing body JSON
	var reqBody struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil 
	{
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido")
		return
	}

	// chiamata al database per aggiornare il nome
	group, err := rt.db.SetGroupName(groupId, reqBody.Name)
	if err != nil 
	{
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore aggiornamento nome")
		return
	}

	// risposta con il gruppo aggiornato e messaggio di successo
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(group)
}

// SetGroupPhoto aggiorna la foto del gruppo (PUT /groups/{groupId}/photo)
func (rt *Router) SetGroupPhoto(w http.ResponseWriter, r *http.Request, userId int, groupId int) 
{

	// Check Auth
	if err := rt.checkAuth(w, r, userId); err != nil { return }

	// verifica che l'utente sia membro del gruppo
	isMember, _ := rt.db.CheckGroupMembership(groupId, userId)
	if !isMember 
	{
		rt.sendError(w, http.StatusForbidden, 403, "Non sei membro del gruppo")
		return
	}

	// Parsing Multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil 
	{
		rt.sendError(w, http.StatusBadRequest, 400, "Errore form dati")
		return
	}

	// estrazione del file
	file, fileHeader, err := r.FormFile("image") // 'image' da YAML
	if err != nil 
	{
		rt.sendError(w, http.StatusBadRequest, 400, "File 'image' mancante")
		return
	}
	defer file.Close()

	// salvataggio su disco 
	storagePath := "./images"
	os.MkdirAll(storagePath, 0755)

	// creazione nome file unico
	filename := fmt.Sprintf("g%d_%d%s", groupId, globaltime.Now().Unix(), filepath.Ext(fileHeader.Filename))
	fullPath := filepath.Join(storagePath, filename)

	// creazione file su disco
	dst, err := os.Create(fullPath)
	if err != nil 
	{
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore server")
		return
	}
	defer dst.Close()
	io.Copy(dst, file)

	// URL relativo per il DB 
	photoURL := "/images/" + filename
	group, err := rt.db.SetGroupPhoto(groupId, photoURL)
	if err != nil 
	{
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore DB")
		return
	}

	// risposta con il gruppo aggiornato e messaggio di successo
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(group)
}