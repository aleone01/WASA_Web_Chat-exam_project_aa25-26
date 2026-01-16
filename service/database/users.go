package database

import (
	"database/sql"
	"errors"
	"strconv"
)

// UserLogin gestisce il meccanismo di accesso o registrazione implicita (Sign-up/Sign-in).
// Flusso logico:
// 1. Tenta di trovare un utente esistente tramite username.
// 2. Se l'utente esiste: aggiorna il suo token di sessione (in questo caso semplice ID convertito) e lo restituisce.
// 3. Se l'utente non esiste: crea un nuovo record nella tabella 'users', genera un token iniziale e lo salva.
// Restituisce l'ID utente, un flag booleano 'created' (true se è un nuovo utente) ed eventuali errori del database.
func (db *appdbimpl) UserLogin(username string) (int, bool, error) {
	var id int

	// Tentativo di recupero ID utente esistente
	err := db.c.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&id)

	if err == nil {
		// Caso: Utente trovato (Login)
		// Aggiornamento del token di sessione (semplificato come stringa dell'ID)
		token := strconv.Itoa(id)

		_, err = db.c.Exec("UPDATE users SET token = ? WHERE id = ?", token, id)
		if err != nil {
			return 0, false, err
		}
		return id, false, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		// Caso: Errore del database diverso da "riga non trovata"
		return 0, false, err
	}

	// Caso: Utente non trovato (Registrazione)
	// Inserimento nuovo record con foto profilo inizialmente NULL
	res, err := db.c.Exec("INSERT INTO users (username, profile_photo, token) VALUES (?, NULL, '')", username)
	if err != nil {
		return 0, false, err
	}

	// Recupero dell'ID autogenerato da SQLite
	lastId, err := res.LastInsertId()
	if err != nil {
		return 0, false, err
	}
	id = int(lastId)

	// Aggiornamento immediato del token per il nuovo utente
	token := strconv.Itoa(id)
	_, err = db.c.Exec("UPDATE users SET token = ? WHERE id = ?", token, id)
	if err != nil {
		return 0, false, err
	}

	return id, true, nil
}

// CheckToken verifica la validità di un token di sessione.
// Viene utilizzato dai middleware o dalle funzioni API per autenticare le richieste HTTP.
// Restituisce l'ID dell'utente associato al token se valido, altrimenti un errore.
func (db *appdbimpl) CheckToken(token string) (int, error) {
	var id int
	err := db.c.QueryRow("SELECT id FROM users WHERE token = ?", token).Scan(&id)
	return id, err
}

// GetUserByUsername recupera i dati pubblici di un utente dato il suo username.
// È fondamentale per la ricerca utenti e per popolare le interfacce utente con nomi e foto.
// Gestisce il caso di foto profilo NULL restituendo uno slice di byte vuoto per evitare panic nel JSON marshalling.
func (db *appdbimpl) GetUserByUsername(username string) (User, error) {

	var u User

	err := db.c.QueryRow("SELECT id, username, profile_photo FROM users WHERE username = ?", username).Scan(&u.Id, &u.Username, &u.ProfilePhoto)
	if err != nil {
		return u, err
	}
	// Normalizzazione del campo BLOB per compatibilità JSON
	if u.ProfilePhoto == nil {
		u.ProfilePhoto = []byte{}
	}
	return u, nil
}

// SetMyUsername permette a un utente di cambiare il proprio nickname.
// Esegue un UPDATE sul database e restituisce la struttura User aggiornata per confermare l'avvenuta modifica al frontend.
func (db *appdbimpl) SetMyUsername(userId int, newUsername string) (User, error) {
	var u User

	// Esecuzione dell'aggiornamento
	_, err := db.c.Exec("UPDATE users SET username = ? WHERE id = ?", newUsername, userId)
	if err != nil {
		return u, err
	}

	// Recupero dei dati aggiornati per la risposta
	err = db.c.QueryRow("SELECT id, username, profile_photo FROM users WHERE id = ?", userId).Scan(&u.Id, &u.Username, &u.ProfilePhoto)
	if u.ProfilePhoto == nil {
		u.ProfilePhoto = []byte{}
	}
	return u, err
}

// SetProfilePhoto aggiorna l'immagine del profilo utente (blob binario).
// Sostituisce l'intero contenuto BLOB nel database per l'utente specificato.
func (db *appdbimpl) SetProfilePhoto(userId int, photoFile []byte) (User, error) {
	var u User

	// Esecuzione dell'aggiornamento del BLOB
	_, err := db.c.Exec("UPDATE users SET profile_photo = ? WHERE id = ?", photoFile, userId)
	if err != nil {
		return u, err
	}

	// Recupero dei dati aggiornati per la risposta
	err = db.c.QueryRow("SELECT id, username, profile_photo FROM users WHERE id = ?", userId).Scan(&u.Id, &u.Username, &u.ProfilePhoto)
	if u.ProfilePhoto == nil {
		u.ProfilePhoto = []byte{}
	}
	return u, err
}
