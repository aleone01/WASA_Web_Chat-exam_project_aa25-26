package api

import 
(
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"fmt"
	"time"
    "os"
)

// SetMyUserName permette all'utente di modificare il proprio username.
func (rt *Router) SetMyUsername(w http.ResponseWriter, r *http.Request, userId int) 
{
	// decodifica della requestBody
	var reqBody SetMyUserNameJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil 
	{
		rt.sendError(w, http.StatusBadRequest, 400, "Formato JSON non valido")
		return
	}

	// verifica dei vincoli sullo username
	if reqBody.Username == nil || len(*reqBody.Username) < 3 || len(*reqBody.Username) > 16 
	{
		rt.sendError(w, http.StatusBadRequest, 400, "Username non valido (lunghezza 3-16 caratteri)")
		return
	}

	newUsername := *reqBody.Username

	// verifica autorizzazione: l'utente loggato deve corrispondere all'ID nel path
	// (Nota: Questo richiede che tu abbia implementato un middleware per estrarre l'utente dalla sessione)
	// if userId != authenticatedUser { ... }

	// chiamata al database per aggiornare l'username
	updatedUser, err := rt.db.SetUsername(userId, newUsername)
	if err != nil 
	{
		// se l'username è già esistente (codice 409)
		if err.Error() == "username already exists" 
		{
			rt.sendError(w, http.StatusConflict, 409, "Username già esistente")
			return
		}

		// caso di errore interno del server
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore interno durante l'aggiornamento username")
		return
	}

	// conversione della struct database 
	response := User
	{
		Id:           &updatedUser.ID,
		Username:     &updatedUser.Username,
		ProfilePhoto: &updatedUser.ProfilePhoto,
	}

	// ritorno del successo
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// SetMyPhoto permette all'utente di aggiornare la propria foto profilo
func (rt *Router) SetMyPhoto(w http.ResponseWriter, r *http.Request, userId int) 
{
	// in caso di errore nel parsing del form
	if err := r.ParseMultipartForm(10 << 20); err != nil 
	{
		rt.sendError(w, http.StatusBadRequest, 400, "Impossibile leggere i dati")
		return
	}

	// recupero dell'immagine dal form 
	file, fileHeader, err := r.FormFile("photo")
	// in caso di errore nel recupero del file
	if err != nil 
	{
		rt.sendError(w, http.StatusBadRequest, 400, "File foto mancante o non valido")
		return
	}
	defer file.Close()

	// lettura dei byte del file
	fileBytes, err := io.ReadAll(file)
	// in caso di errore nella lettura del file
	if err != nil 
	{
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore nella lettura del file")
		return
	}

	// controllo del file 
	if len(fileBytes) == 0 
	{
		rt.sendError(w, http.StatusBadRequest, 400, "Il file inviato è vuoto")
		return
	}

	// creazione di una cartella di storage, se non esiste
	storagePath := "./images"
	// in caso di errore nella creazione della cartella
    if err := os.MkdirAll(storagePath, os.ModePerm); err != nil 
    {
        rt.sendError(w, http.StatusInternalServerError, 500, "Errore interno del server")
        return
    }

    // creazione di un nome file univoco, per evitare sovrascritture e conflitti
    filename := fmt.Sprintf("user_%d_%d%s", userId, time.Now().Unix(), filepath.Ext(fileHeader.Filename))
    fullPath := filepath.Join(storagePath, filename)

    // creazione del file fisico sul disco del server
    destinationFile, err := os.Create(fullPath)
	// in caso di errore nella creazione del file
    if err != nil 
    {
        rt.sendError(w, http.StatusInternalServerError, 500, "Impossibile salvare il file")
        return
    }
    defer destinationFile.Close()

    // copia dello stream di byte dal file caricato al file di destinazione
    if _, err := io.Copy(destinationFile, file); err != nil 
    {
        rt.sendError(w, http.StatusInternalServerError, 500, "Errore durante la scrittura del file")
        return
    }
	
    // passaggio dei byte o l'url al DB
    photoURL := fmt.Sprintf("https://example.com/images/%s", filename)

	// chiamata al database per aggiornare la foto
	updatedUser, err := rt.db.SetProfilePhoto(userId, photoURL)
	// in caso di errore nel database
	if err != nil 
	{
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore interno durante l'aggiornamento foto")
		return
	}

	// conversione della struct database 
	response := User{
		Id:           &updatedUser.ID,
		Username:     &updatedUser.Username,
		ProfilePhoto: &updatedUser.ProfilePhoto,
	}

	// ritorno del successo
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}