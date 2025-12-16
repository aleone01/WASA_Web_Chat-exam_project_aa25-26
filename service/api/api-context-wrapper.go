package api

import (
	"net/http"

	"github.com/aleone01/Web-Project-repo/service/api/reqcontext"
	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

// httpRouterHandler definisce la firma per le funzioni handler che richiedono, oltre ai parametri standard
// di httprouter, anche un'istanza di reqcontext.RequestContext per gestire i dati specifici della richiesta.
type httpRouterHandler func(http.ResponseWriter, *http.Request, httprouter.Params, reqcontext.RequestContext)

// wrap funge da middleware per gli handler delle rotte. Questa funzione crea un nuovo RequestContext per ogni
// chiamata, generando un UUID univoco per la richiesta e inizializzando un logger strutturato con i campi
// identificativi (reqid e remote-ip). Successivamente, passa questo contesto alla funzione handler specificata.
func (rt *_router) wrap(fn httpRouterHandler) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		reqUUID, err := uuid.NewV4()
		if err != nil {
			rt.baseLogger.WithError(err).Error("can't generate a request UUID")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var ctx = reqcontext.RequestContext{
			ReqUUID: reqUUID,
		}

		ctx.Logger = rt.baseLogger.WithFields(logrus.Fields{
			"reqid":     ctx.ReqUUID.String(),
			"remote-ip": r.RemoteAddr,
		})

		fn(w, r, ps, ctx)
	}
}
