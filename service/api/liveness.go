package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// liveness è un handler HTTP progettato per verificare lo stato di salute del server API (Health Check).
// Il suo scopo principale è rispondere con uno stato HTTP 200 OK se il server è attivo e in grado di ricevere richieste.
// In uno scenario di produzione più complesso, questa funzione potrebbe verificare anche la connessione al database
// o ad altri servizi critici, restituendo un errore 500 in caso di malfunzionamenti.
func (rt *_router) liveness(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
}
