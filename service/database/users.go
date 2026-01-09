package database

import (
	"database/sql"
	"errors"
	"strconv"
)

// UserLogin gestisce l'accesso o la registrazione di un utente.
// Cerca l'utente nel database tramite username: se esiste, aggiorna il token di sessione;
// se non esiste, crea un nuovo record utente e genera un token iniziale.
// Restituisce l'ID utente, un booleano che indica se è stato creato un nuovo utente, ed eventuali errori.
func (db *appdbimpl) UserLogin(username string) (int, bool, error) {
	var id int

	err := db.c.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&id)

	if err == nil {
		token := strconv.Itoa(id)

		_, err = db.c.Exec("UPDATE users SET token = ? WHERE id = ?", token, id)
		if err != nil {
			return 0, false, err
		}
		return id, false, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return 0, false, err
	}

	res, err := db.c.Exec("INSERT INTO users (username, profile_photo, token) VALUES (?, NULL, '')", username)
	if err != nil {
		return 0, false, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return 0, false, err
	}
	id = int(lastId)

	token := strconv.Itoa(id)
	_, err = db.c.Exec("UPDATE users SET token = ? WHERE id = ?", token, id)
	if err != nil {
		return 0, false, err
	}

	return id, true, nil
}

// CheckToken valida un token di autenticazione confrontandolo con quelli salvati nel database.
// Se il token è valido, restituisce l'ID dell'utente associato.
func (db *appdbimpl) CheckToken(token string) (int, error) {
	var id int
	err := db.c.QueryRow("SELECT id FROM users WHERE token = ?", token).Scan(&id)
	return id, err
}

// GetUserByUsername recupera le informazioni di un utente (ID, username, foto) cercandolo per nome utente.
// Utile per risolvere riferimenti utente nelle operazioni di chat o ricerca.
func (db *appdbimpl) GetUserByUsername(username string) (User, error) {

	var u User

	err := db.c.QueryRow("SELECT id, username, profile_photo FROM users WHERE username = ?", username).Scan(&u.Id, &u.Username, &u.ProfilePhoto)
	if err != nil {
		return u, err
	}
	if u.ProfilePhoto == nil {
		u.ProfilePhoto = []byte{}
	}
	return u, nil
}

// SetMyUsername aggiorna il nome utente per un dato ID.
// Esegue l'update nel database e restituisce la struttura User aggiornata per conferma.
func (db *appdbimpl) SetMyUsername(userId int, newUsername string) (User, error) {
	var u User

	_, err := db.c.Exec("UPDATE users SET username = ? WHERE id = ?", newUsername, userId)
	if err != nil {
		return u, err
	}

	err = db.c.QueryRow("SELECT id, username, profile_photo FROM users WHERE id = ?", userId).Scan(&u.Id, &u.Username, &u.ProfilePhoto)
	if u.ProfilePhoto == nil {
		u.ProfilePhoto = []byte{}
	}
	return u, err
}

// SetProfilePhoto aggiorna l'URL della foto profilo di un utente.
// Simile a SetMyUsername, esegue l'update e restituisce l'oggetto User aggiornato.
func (db *appdbimpl) SetProfilePhoto(userId int, photoFile []byte) (User, error) {
	var u User

	_, err := db.c.Exec("UPDATE users SET profile_photo = ? WHERE id = ?", photoFile, userId)
	if err != nil {
		return u, err
	}

	err = db.c.QueryRow("SELECT id, username, profile_photo FROM users WHERE id = ?", userId).Scan(&u.Id, &u.Username, &u.ProfilePhoto)
	if u.ProfilePhoto == nil {
		u.ProfilePhoto = []byte{}
	}
	return u, err
}
