package api

import (
	"encoding/json"
	"net/http"

	"github.com/aleone01/Web-Project-repo/service/database"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

// Router è l'interfaccia pubblica che espone le funzionalità principali del componente API,
// ovvero la capacità di restituire un handler HTTP per le richieste e un metodo per chiudere le risorse.
type Router interface {
	Handler() http.Handler
	Close() error
}

// _router è la struttura interna che implementa l'interfaccia Router. Essa mantiene lo stato necessario
// per il funzionamento dell'API, includendo il router HTTP effettivo, la connessione al database e il logger base.
type _router struct {
	router     *httprouter.Router
	db         database.AppDatabase
	baseLogger *logrus.Logger
}

// Config è una struttura utilizzata per passare le dipendenze necessarie durante la creazione di una nuova istanza
// dell'API, in particolare il logger per il tracciamento e l'interfaccia di accesso al database.
type Config struct {
	Logger   *logrus.Logger
	Database database.AppDatabase
}

// New inizializza e restituisce una nuova istanza del router API. Utilizza la configurazione fornita per
// impostare il router HTTP di base, collegare il database e configurare il sistema di logging.
func New(cfg Config) (Router, error) {
	rt := &_router{
		router:     httprouter.New(),
		db:         cfg.Database,
		baseLogger: cfg.Logger,
	}

	return rt, nil
}

// checkAuth verifica l'autenticazione dell'utente per una data richiesta. Estrae il token dall'header
// (supportando sia il formato "Bearer" che "sessionToken"), controlla la sua validità nel database e
// si assicura che l'ID utente associato al token corrisponda all'ID richiesto per l'operazione.
// In caso di errore, invia una risposta HTTP appropriata e restituisce false.
func (rt *_router) checkAuth(w http.ResponseWriter, r *http.Request, requiredUserId int) bool {

	token := r.Header.Get("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	} else if token == "" {
		token = r.Header.Get("sessionToken")
	}

	userIdFromToken, err := rt.db.CheckToken(token)
	if err != nil {
		rt.sendError(w, http.StatusUnauthorized, 401, "Token non valido o utente non trovato")
		return false
	}

	if userIdFromToken != requiredUserId {
		rt.sendError(w, http.StatusForbidden, 403, "Accesso negato alla risorsa richiesta")
		return false
	}

	return true
}

// sendError è una funzione di utilità per inviare risposte di errore in formato JSON al client.
// Imposta l'header Content-Type, scrive il codice di stato HTTP e serializza un oggetto JSON contenente
// un codice di errore interno e un messaggio descrittivo.
func (rt *_router) sendError(w http.ResponseWriter, status int, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{
		Code:    code,
		Message: message,
	})
}
