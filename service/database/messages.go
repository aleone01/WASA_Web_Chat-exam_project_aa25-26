package database

import (
	"database/sql"
	"errors"
	"time"
)

// CreateMessage inserisce un nuovo messaggio nel database.
// Gestisce parametri opzionali come 'replyTo' (per le risposte) e 'isForward' (per i messaggi inoltrati).
// Dopo l'inserimento, costruisce e restituisce l'oggetto Message completo con l'ID generato.
func (db *appdbimpl) CreateMessage(chatId int, userId int, text string, photoFile []byte, sentAt time.Time, replyTo int, isForward bool) (Message, error) {
	var m Message

	// Query di inserimento con gestione dei campi opzionali
	query := `INSERT INTO messages (chat_id, user_id, text, photo_file, sent_at, is_read, reply_to_message_id, is_forward) VALUES (?, ?, ?, ?, ?, FALSE, ?, ?)`

	// Gestione del campo opzionale replyTo
	var replyToVal sql.NullInt64
	if replyTo > 0 {
		replyToVal.Int64 = int64(replyTo)
		replyToVal.Valid = true
	}

	// Esecuzione dell'inserimento
	res, err := db.c.Exec(query, chatId, userId, text, photoFile, sentAt, replyToVal, isForward)
	if err != nil {
		return m, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return m, err
	}

	// Costruzione dell'oggetto Message da restituire
	m.Id = int(lastId)
	m.ChatId = chatId
	m.SentBy = userId
	m.Text = text
	m.PhotoFile = photoFile
	m.SentAt = sentAt
	m.Checkmark = false
	m.ReplyTo = replyTo
	m.IsForward = isForward

	return m, nil
}

// AddReaction salva una reazione (emoticon) di un utente a uno specifico messaggio.
// L'unicità della coppia messaggio-utente è garantita dalla chiave primaria della tabella reactions.
func (db *appdbimpl) AddReaction(messageId int, userId int, emoticon string) error {
	_, err := db.c.Exec(`
		INSERT INTO reactions (message_id, user_id, emoticon) 
		VALUES (?, ?, ?)`,
		messageId, userId, emoticon)

	return err
}

// RemoveReaction elimina una reazione precedentemente aggiunta da un utente a un messaggio.
func (db *appdbimpl) RemoveReaction(messageId int, userId int) error {
	_, err := db.c.Exec("DELETE FROM reactions WHERE message_id = ? AND user_id = ?", messageId, userId)
	return err
}

// DeleteMessage cancella un messaggio dal database, ma solo se l'utente richiedente ne è l'autore.
// Restituisce un errore specifico se il messaggio non esiste o se l'utente non è autorizzato.
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

// GetMessage recupera i dettagli di un singolo messaggio tramite il suo ID.
// Restituisce una struttura Message contenente i dati essenziali (testo, foto, timestamp, ecc.).
func (db *appdbimpl) GetMessage(messageId int) (Message, error) {
	var m Message

	err := db.c.QueryRow(`
		SELECT id, chat_id, user_id, text, photo_file, sent_at 
		FROM messages WHERE id = ?`,
		messageId).Scan(&m.Id, &m.ChatId, &m.SentBy, &m.Text, &m.PhotoFile, &m.SentAt)

	if err != nil {
		return m, err
	}

	if m.PhotoFile == nil {
		m.PhotoFile = []byte{}
	}
	m.Checkmark = true
	return m, nil
}
