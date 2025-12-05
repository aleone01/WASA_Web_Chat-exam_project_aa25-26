package database

import (
	"errors"
	"time"
)

// CreateMessage inserisce un nuovo messaggio nel database
func (db *appdbimpl) CreateMessage(chatId int, userId int, text string, photoUrl string, sentAt time.Time) (Message, error) {
	var m Message
	m.ChatId = chatId
	m.SentBy = userId
	m.Text = text
	m.PhotoUrl = photoUrl
	m.SentAt = sentAt
	m.Checkmark = true

	// Esecuzione INSERT
	res, err := db.c.Exec(`
		INSERT INTO messages (chat_id, user_id, text, photo_url, sent_at) 
		VALUES (?, ?, ?, ?, ?)`,
		chatId, userId, text, photoUrl, m.SentAt)

	if err != nil {
		return m, err
	}

	// Recupero ID generato
	lastId, err := res.LastInsertId()
	if err != nil {
		return m, err
	}
	m.Id = int(lastId)

	return m, nil
}

// AddReaction aggiunge un'emoticon a un messaggio (È QUESTA CHE MANCAVA)
func (db *appdbimpl) AddReaction(messageId int, userId int, emoticon string) error {
	_, err := db.c.Exec(`
		INSERT INTO reactions (message_id, user_id, emoticon) 
		VALUES (?, ?, ?)`,
		messageId, userId, emoticon)

	return err
}

// RemoveReaction rimuove la reazione dell'utente da un messaggio
func (db *appdbimpl) RemoveReaction(messageId int, userId int) error {
	_, err := db.c.Exec("DELETE FROM reactions WHERE message_id = ? AND user_id = ?", messageId, userId)
	return err
}

// DeleteMessage elimina un messaggio (solo se l'utente ne è il proprietario)
func (db *appdbimpl) DeleteMessage(messageId int, userId int) error {
	res, err := db.c.Exec("DELETE FROM messages WHERE id = ? AND user_id = ?", messageId, userId)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("message not found or unauthorized")
	}

	return nil
}

// GetMessage recupera un singolo messaggio per ID (utile per il forward)
func (db *appdbimpl) GetMessage(messageId int) (Message, error) {
	var m Message
	err := db.c.QueryRow(`
		SELECT id, chat_id, user_id, text, COALESCE(photo_url, ''), sent_at 
		FROM messages WHERE id = ?`,
		messageId).Scan(&m.Id, &m.ChatId, &m.SentBy, &m.Text, &m.PhotoUrl, &m.SentAt)

	if err != nil {
		return m, err
	}
	m.Checkmark = true
	return m, nil
}
