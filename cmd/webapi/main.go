package main

import (
	"database/sql"
	"errors"
	"net/http"
	"os"

	"github.com/aleone01/Web-Project-repo/service/api"
	"github.com/aleone01/Web-Project-repo/service/database"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

func main() {

	// configura logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	// connessione al database SQLite
	dbPath := "./wasatext.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		logger.WithError(err).Fatal("Errore apertura connessione SQL")
	}
	defer db.Close()

	// inizializza tabelle database
	appDB, err := database.New(db)
	if err != nil {
		logger.WithError(err).Fatal("Errore inizializzazione tabelle database")
	}
	logger.Info("Database connesso e inizializzato correttamente")

	// crea API handler
	apiHandler, err := api.New(api.Config{
		Logger:   logger,
		Database: appDB,
	})
	if err != nil {
		logger.WithError(err).Fatal("Errore creazione API Handler")
	}

	// crea cartella per immagini se non esiste
	imagesDir := "./images"
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		logger.WithError(err).Fatal("Impossibile creare cartella images")
	}

	// crea router principale
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir(imagesDir))
	mux.Handle("/images/", http.StripPrefix("/images/", fs))
	mux.Handle("/", apiHandler)

	// applica CORS a tutti gli handler
	globalHandler := applyCORSHandler(mux)

	// avvio server HTTP
	serverPort := ":3000"
	logger.Infof("Server API avviato sulla porta %s", serverPort)

	// avvio server
	if err := http.ListenAndServe(serverPort, globalHandler); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			logger.WithError(err).Fatal("Errore imprevisto del server HTTP")
		}
	}
}
