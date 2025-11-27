package database

import (
	"database/sql"
	"errors"
)

// GetConversations restituisce la lista delle chat
func (db *appdbimpl) GetConversations(userId int) ([]ChatListItem, error) {

	var chats []ChatListItem

	// query standard per ottenere le chat dell'utente
	query := `
		SELECT 
			c.id, 
			c.is_group, 
			COALESCE(c.group_photo, ''), 
			MAX(m.sent_at) as last_msg_time,
			COALESCE(m.text, '') as snippet
		FROM chats c
		JOIN members mem ON c.id = mem.chat_id
		LEFT JOIN messages m ON c.id = m.chat_id
		WHERE mem.user_id = ?
		GROUP BY c.id
		ORDER BY last_msg_time DESC
	`
	// esecuzione della query
	rows, err := db.c.Query(query, userId)
	// in caso di errore
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// iterazione sui risultati
	for rows.Next() {
		var c ChatListItem
		c.SnippetIcon = "" // default vuoto

		// scan dei valori nelle variabili
		if err := rows.Scan(&c.Id, &c.IsGroup, &c.PhotoChat, &c.LastMessage, &c.SnippetText); err != nil {
			continue // salta righe con errori di scan
		}
		chats = append(chats, c) //
	}

	return chats, nil
}

// GetChatWithUser cerca una chat privata con un altro utente
func (db *appdbimpl) GetChatWithUser(userId int, targetUsername string) ([]ChatListItem, error) {

	var chats []ChatListItem

	// ottengiene l'ID dell'altro utente
	var targetId int
	err := db.c.QueryRow("SELECT id FROM users WHERE username = ?", targetUsername).Scan(&targetId)
	if err != nil {
		return chats, nil // utente non trovato -> lista vuota
	}

	// cerca la chat in comune NON di gruppo
	query := `
		SELECT c.id, c.is_group, COALESCE(c.group_photo, '')
		FROM chats c
		JOIN members m1 ON c.id = m1.chat_id
		JOIN members m2 ON c.id = m2.chat_id
		WHERE m1.user_id = ? AND m2.user_id = ? AND c.is_group = FALSE
	`
	// esecuzione della query
	rows, err := db.c.Query(query, userId, targetId)
	// in caso di errore
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// iterazione sui risultati
	for rows.Next() {
		var c ChatListItem
		c.SnippetText = ""
		// scan dei valori nelle variabili
		if err := rows.Scan(&c.Id, &c.IsGroup, &c.PhotoChat); err != nil {
			return nil, err
		}
		chats = append(chats, c)
	}

	return chats, nil
}

// GetChatMessages restituisce i messaggi di una chat
func (db *appdbimpl) GetChatMessages(chatId int, userId int) ([]Message, error) {

	// controlla se l'utente è membro della chat
	var exists int
	err := db.c.QueryRow("SELECT 1 FROM members WHERE chat_id = ? AND user_id = ?", chatId, userId).Scan(&exists)
	if err == sql.ErrNoRows {
		return nil, errors.New("chat non trovata o accesso negato")
	} else if err != nil {
		return nil, err
	}

	// ottiene i messaggi della chat
	query := `
		SELECT id, text, sent_at, user_id, COALESCE(photo_url, '')
		FROM messages
		WHERE chat_id = ?
		ORDER BY sent_at DESC
	`
	// esecuzione della query
	rows, err := db.c.Query(query, chatId)
	// in caso di errore
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// iterazione sui risultati
	var messages []Message
	for rows.Next() {
		var m Message
		// scan dei valori nelle variabili
		if err := rows.Scan(&m.Id, &m.Text, &m.SentAt, &m.SentBy, &m.PhotoUrl); err != nil {
			return nil, err
		}
		m.Checkmark = true // tutti i messaggi sono considerati letti
		messages = append(messages, m)
	}

	return messages, nil
}
