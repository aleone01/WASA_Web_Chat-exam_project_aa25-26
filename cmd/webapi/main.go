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

	// Inizializzazione Tabelle (Service Database)
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

	// Configurazione Static Files (Immagini)
	imagesDir := "./images"
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		logger.WithError(err).Fatal("Impossibile creare cartella images")
	}

	// Router Principale (ServeMux)
	// Combiniamo API e FileServer
	mux := http.NewServeMux()

	// Rotta per le immagini: http://localhost:3000/images/nomefile.jpg
	fs := http.FileServer(http.Dir(imagesDir))
	mux.Handle("/images/", http.StripPrefix("/images/", fs))

	// Rotta per l'API
	mux.Handle("/", apiHandler)

	// Applicazione CORS (usando il file cors.go)
	// Avvolgiamo il mux con la funzione definita in cors.go
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
