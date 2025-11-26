package api

import 
(
	"encoding/json"
	"net/http"
)

// DoLogin gestisce l'autenticazione: verifica se l'utente esiste: se sì lo logga (200), se no lo crea (201).
func (rt *Router) DoLogin(w http.ResponseWriter, r *http.Request) 
{
	// decodifica della requestBody
	var reqBody DoLoginJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil 
	{
		// Se il JSON non è valido, restituisce un errore 400
		rt.sendError(w, http.StatusBadRequest, 400, "Formato JSON non valido")
		return
	}

	// verifica dei vincoli sullo username
	if reqBody.Username == nil || len(*reqBody.Username) < 3 || len(*reqBody.Username) > 16 
	{
		rt.sendError(w, http.StatusBadRequest, 400, "Username non valido (lunghezza 3-16 caratteri)")
		return
	}

	// estrazione dei parametri
	username := *reqBody.Username

	// chiamata al database che restituisce: id, created (bool), error
	userId, created, err := rt.db.UserLogin(username)
	if err != nil 
	{
		// caso di errore interno del server
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore interno durante il login")
		return
	}

	// costruzione dell'oggetto risposta
	response := LoginResponse{
		Identifier: &userId,
	}

	// 5. Invio risposta
	w.Header().Set("Content-Type", "application/json")

	// Ritorno del codice di successo: 200 -> utente loggato, 201 -> utente creato
	if created 
	{
		w.WriteHeader(http.StatusCreated) // 201 Created
	} 
	else 
	{
		w.WriteHeader(http.StatusOK) // 200 OK
	}

	// Scrittura del JSON finale
	json.NewEncoder(w).Encode(response)
}