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
	"time"
)

// SetMyUserName gestisce il cambio username (PUT /users/{userId}/username)
func (rt *Router) SetMyUserName(w http.ResponseWriter, r *http.Request, userId int) {
	
	// controlla che l'utente sia  se stesso
	if err := rt.checkAuth(w, r, userId); err != nil 
	{
		return
	}

	// parsing JSON
	var reqBody struct 
	{
		Username *string `json:"username"`
	}

	// in caso di errore nel decoding
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido")
		return
	}

	// in caso di username non valido
	if reqBody.Username == nil || len(*reqBody.Username) < 3 || len(*reqBody.Username) > 16 {
		rt.sendError(w, http.StatusBadRequest, 400, "Username non valido (3-16 caratteri)")
		return
	}

	// chiamata al database per aggiornare l'username
	user, err := rt.db.SetMyUsername(userId, *reqBody.Username)
	// in caso di errore nell'aggiornamento
	if err != nil 
	{
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore aggiornamento username")
		return
	}

	// risposta con il nuovo profilo utente
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// SetMyPhoto gestisce il cambio foto (PUT /users/{userId}/photo)
func (rt *Router) SetMyPhoto(w http.ResponseWriter, r *http.Request, userId int) 
{
	
	// Check Auth
	if err := rt.checkAuth(w, r, userId); err != nil 
	{
		return
	}

	// parsing Multipart 
	if err := r.ParseMultipartForm(10 << 20); err != nil 
	{
		rt.sendError(w, http.StatusBadRequest, 400, "File troppo grande o errore form")
		return
	}

	// estrazione del File
	file, fileHeader, err := r.FormFile("photo")
	if err != nil 
	{
		rt.sendError(w, http.StatusBadRequest, 400, "Campo 'photo' mancante")
		return
	}
	defer file.Close()

	// salvataggio su disco
	storagePath := "./images"
	os.MkdirAll(storagePath, 0755) // crea la cartella se non esiste

	// nome unico per il file
	filename := fmt.Sprintf("u%d_%d%s", userId, globaltime.Now().Unix(), filepath.Ext(fileHeader.Filename))
	fullPath := filepath.Join(storagePath, filename)

	// creazione del file sul disco
	dst, err := os.Create(fullPath)
	if err != nil 
	{
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore creazione file")
		return
	}
	defer dst.Close()

	// copia del contenuto
	if _, err := io.Copy(dst, file); err != nil 
	{
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore salvataggio file")
		return
	}

	// chiamata al database per aggiornare la foto profilo
	photoURL := "/images/" + filename
	user, err := rt.db.SetProfilePhoto(userId, photoURL)
	if err != nil 
	{
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore DB")
		return
	}

	// risposta con il nuovo profilo utente
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}