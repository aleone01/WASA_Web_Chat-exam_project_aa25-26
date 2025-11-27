package database

import (
	"database/sql"
	"errors"
)

// UserLogin verifica se l'utente esiste (200) o lo crea (201)
func (db *appdbimpl) UserLogin(username string) (int, bool, error) {
	var id int

	// prova a leggere l'ID utente
	err := db.c.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&id)

	if err == nil {
		// utente trovato -> restituisco ID e false (non creato nuovo)
		return id, false, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		// errore generico database
		return 0, false, err
	}

	// se non trovato, eseguo INSERT per creare nuovo utente
	res, err := db.c.Exec("INSERT INTO users (username, profile_photo) VALUES (?, '')", username)
	// in caso di errore
	if err != nil {
		return 0, false, err
	}

	// recupero l'ID generato
	lastId, err := res.LastInsertId()
	// in caso di errore
	if err != nil {
		return 0, false, err
	}

	// restituisco l'ID del nuovo utente e true (nuovo utente creato)
	return int(lastId), true, nil
}

// CheckToken verifica se il token è valido e restituisce l'userId associato
func (db *appdbimpl) CheckToken(token string) (int, error) {
	var id int
	err := db.c.QueryRow("SELECT id FROM users WHERE token = ?", token).Scan(&id)
	return id, err
}

// SetMyUsername aggiorna lo username e restituisce l'oggetto User aggiornato
func (db *appdbimpl) SetMyUsername(userId int, newUsername string) (User, error) {
	var u User

	// eseguo l'aggiornamento
	_, err := db.c.Exec("UPDATE users SET username = ? WHERE id = ?", newUsername, userId)
	// in caso di errore
	if err != nil {
		return u, err
	}

	// legge i dati aggiornati per restituirli
	err = db.c.QueryRow("SELECT id, username, profile_photo FROM users WHERE id = ?", userId).Scan(&u.Id, &u.Username, &u.ProfilePhoto)
	return u, err
}

// SetProfilePhoto aggiorna la foto e restituisce l'oggetto User aggiornato
func (db *appdbimpl) SetProfilePhoto(userId int, photoURL string) (User, error) {
	var u User

	// eseguo l'aggiornamento
	_, err := db.c.Exec("UPDATE users SET profile_photo = ? WHERE id = ?", photoURL, userId)
	// in caso di errore
	if err != nil {
		return u, err
	}

	// legge i dati aggiornati per restituirli
	err = db.c.QueryRow("SELECT id, username, profile_photo FROM users WHERE id = ?", userId).Scan(&u.Id, &u.Username, &u.ProfilePhoto)
	return u, err
}
