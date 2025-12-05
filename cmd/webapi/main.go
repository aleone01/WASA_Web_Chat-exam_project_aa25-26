package main

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/aleone01/Web-Project-repo/service/api"
	"github.com/aleone01/Web-Project-repo/service/database"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

func main() {
	// Inizializzazione Logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	// Apertura Connessione Database
	dbPath := "./wasatext.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		logger.WithError(err).Fatal("Errore apertura connessione SQL")
	}
	defer db.Close()

	// Inizializzazione Tabelle
	appDB, err := database.New(db)
	if err != nil {
		logger.WithError(err).Fatal("Errore inizializzazione tabelle database")
	}
	logger.Info("Database connesso e inizializzato correttamente")

	// Inizializzazione API Router
	apiHandler, err := api.New(api.Config{
		Logger:   logger,
		Database: appDB,
	})
	if err != nil {
		logger.WithError(err).Fatal("Errore creazione API Handler")
	}

	// Configurazione Router
	mux := http.NewServeMux()

	// Registrazione Handler API
	mux.Handle("/", apiHandler)

	// Applicazione CORS
	globalHandler := applyCORSHandler(mux)

	// Avvio Server
	serverPort := ":3000"
	logger.Infof("Server API in ascolto su http://localhost%s", serverPort)

	if err := http.ListenAndServe(serverPort, globalHandler); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			logger.WithError(err).Fatal("Errore imprevisto del server HTTP")
		}
	}
}
