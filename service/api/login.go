package api

// Import del pacchetto contenente le struct generate (se sono nello stesso package 'api', non serve import)
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
		//se il JSON non è valido, restituisce un errore 400
		rt.sendError(w, http.StatusBadRequest, 400, "Formato non valido")
		return
	}

	// verifica dei vincoli sullo username
	if reqBody.Username == nil || len(*reqBody.Username) < 3 || len(*reqBody.Username) > 16 
	{
		rt.sendError(w, http.StatusBadRequest, 400, "Username non valido (lunghezza 3-16 caratteri)")
		return
	}

	username := *reqBody.Username

	// chiamata al database
	userId, userCreated, err := rt.db.UserLogin(username)
	if err != nil 
	{
    	// caso di errore interno del server
    	rt.sendError(w, http.StatusInternalServerError, 500, "Errore interno durante il login")
    	return
	}

	// risposta del login
	response := LoginResponse
	{
    	Identifier: &userId,
	}

	// caso di errore interno del server
	if err != nil {
		
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore interno durante il login")
		return
	}

	// costruzione dell'oggetto LoginResponse
	response := LoginResponse
	{
		Identifier: &userId,
	}

	// ritorno del codice di successo: 200 -> utente loggato, 201 -> utente creato
	if created 
	{
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated) // 201
		json.NewEncoder(w).Encode(response)
	} 
	else 
	{
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // 200
		json.NewEncoder(w).Encode(response)
	}
}