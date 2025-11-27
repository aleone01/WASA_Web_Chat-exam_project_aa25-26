package api

import (
	"encoding/json"
	"net/http"

	"github.com/aleone01/Web-Project-repo/service/database"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

// _router è l'implementazione interna del router API.
// Questa struct sostituisce l'interfaccia Router che causava errori "undefined".
type _router struct {
	router     *httprouter.Router
	db         database.AppDatabase
	baseLogger *logrus.Logger
}

// Config è la struttura di configurazione per creare una nuova API
type Config struct {
	Logger   *logrus.Logger
	Database database.AppDatabase
}

// New crea e restituisce un nuovo http.Handler (l'API pronta all'uso).
func New(cfg Config) (http.Handler, error) {
	// Inizializza la struct privata _router
	rt := &_router{
		router:     httprouter.New(),
		db:         cfg.Database,
		baseLogger: cfg.Logger,
	}

	// Chiama Handler() (che si trova in api-handler.go) per registrare tutte le rotte
	return rt.Handler(), nil
}

// checkAuth controlla se il token nell'header è valido e corrisponde all'userId richiesto
func (rt *_router) checkAuth(w http.ResponseWriter, r *http.Request, requiredUserId int) bool {
	// estrazione token
	token := r.Header.Get("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	} else if token == "" {
		token = r.Header.Get("sessionToken")
	}

	// verifica database
	userIdFromToken, err := rt.db.CheckToken(token)
	if err != nil {
		rt.sendError(w, http.StatusUnauthorized, 401, "Token non valido o utente non trovato")
		return false
	}

	// verifica Identità
	if userIdFromToken != requiredUserId {
		rt.sendError(w, http.StatusForbidden, 403, "Accesso negato alla risorsa richiesta")
		return false
	}

	return true
}

// sendError è l'helper per le risposte JSON di errore
func (rt *_router) sendError(w http.ResponseWriter, status int, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// invio risposta JSON
	_ = json.NewEncoder(w).Encode(struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{
		Code:    code,
		Message: message,
	})
}
