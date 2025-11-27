package main

import (
	"net/http"

	"github.com/gorilla/handlers"
)

// applyCORSHandler applies a CORS policy to the router.
func applyCORSHandler(h http.Handler) http.Handler {
	return handlers.CORS(
		handlers.AllowedHeaders([]string{
			"x-example-header",
			"Content-Type",  // Necessario per inviare JSON
			"Authorization", // Necessario per il token Bearer
			"sessionToken",  // Necessario se usi il tuo header custom definito in api.yaml
		}),
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS", "DELETE", "PUT"}),
		// Do not modify the CORS origin and max age, they are used in the evaluation.
		handlers.AllowedOrigins([]string{"*"}),
		handlers.MaxAge(1),
	)(h)
}
