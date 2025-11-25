package api

import 
(
	"context"
	"net/http"
	"strings"
)

// initd della chiave unica per evitare collisioni
type contextKey string

const 
(
	// ContextKeyUser è la chiave usata per salvare l'ID utente nel context della richiesta
	ContextKeyUser contextKey = "userID"
)

// AuthMiddleware verifica l'autenticazione dell'utente tramite header 'sessionToken'
func (rt *Router) AuthMiddleware(next http.Handler) http.Handler 
{
	return http.HandlerFunc(func(w, r, r2) 
	{ 
		
		// estrazione del token dall'header
		token := r.Header.Get("sessionToken")

		// standard "Authorization: Bearer <token>" 
		if token == "" {
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		// se non c'è token -> errore 401 
		if token == "" {
			rt.sendError(w, http.StatusUnauthorized, 401, "Autenticazione richiesta (token mancante)")
			return
		}

		// validazione del token tramite Database
		userID, err := rt.db.CheckToken(token)
		if err != nil {
			// il caso in cui il DB dice che il token non esiste o è scaduto
			rt.sendError(w, http.StatusUnauthorized, 401, "Token non valido o scaduto")
			return
		}

		// inserimento dell'ID utente nel contesto della richiesta
		ctx := context.WithValue(r.Context(), ContextKeyUser, userID)

		// passaggio del controllo al prossimo handler con il nuovo contesto
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}