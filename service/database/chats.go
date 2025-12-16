package database

import (
	"database/sql"
)

// GetMyConversations recupera l'elenco di tutte le conversazioni (sia gruppi che chat private) a cui partecipa l'utente.
// Per ogni chat, restituisce l'ultimo messaggio scambiato (testo o foto), il timestamp e le informazioni sull'interlocutore o sul gruppo.
// I risultati sono ordinati cronologicamente in base all'ultimo messaggio.
func (db *appdbimpl) GetMyConversations(userId int) ([]ChatListItem, error) {

	chats := make([]ChatListItem, 0)

	query := `
		SELECT 
			c.id, 
			c.is_group, 
			COALESCE(c.group_name, ''),
			COALESCE(c.group_photo, ''),
			(SELECT sent_at FROM messages WHERE chat_id = c.id ORDER BY sent_at DESC LIMIT 1) as last_msg_time,
			(SELECT text FROM messages WHERE chat_id = c.id ORDER BY sent_at DESC LIMIT 1) as snippet_text,
			(SELECT photo_url FROM messages WHERE chat_id = c.id ORDER BY sent_at DESC LIMIT 1) as snippet_photo,
			COALESCE(u.username, ''),
			COALESCE(u.profile_photo, '')
		FROM chats c
		JOIN members m1 ON c.id = m1.chat_id
		LEFT JOIN members m2 ON c.id = m2.chat_id AND m2.user_id != m1.user_id
		LEFT JOIN users u ON m2.user_id = u.id
		WHERE m1.user_id = ?
		GROUP BY c.id
		ORDER BY last_msg_time DESC
	`

	rows, err := db.c.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c ChatListItem
		var groupName, groupPhoto string
		var lastTime sql.NullTime
		var lastText, lastPhoto, otherUsername, otherPhoto sql.NullString

		c.SnippetIcon = ""

		if err := rows.Scan(&c.Id, &c.IsGroup, &groupName, &groupPhoto, &lastTime, &lastText, &lastPhoto, &otherUsername, &otherPhoto); err != nil {
			continue
		}

		if lastTime.Valid {
			c.LastMessage = lastTime.Time
		}

		if lastText.Valid && lastText.String != "" {
			c.SnippetText = lastText.String
		} else if lastPhoto.Valid && lastPhoto.String != "" {
			c.SnippetText = "📷 Foto"
		} else {
			c.SnippetText = ""
		}

		if c.IsGroup {
			c.Name = groupName
			c.PhotoChat = groupPhoto
		} else {
			if otherUsername.Valid {
				c.Name = otherUsername.String
			} else {
				c.Name = "Utente Sconosciuto"
			}
			if otherPhoto.Valid {
				c.PhotoChat = otherPhoto.String
			} else {
				c.PhotoChat = ""
			}
		}

		chats = append(chats, c)
	}

	return chats, nil
}

// GetChatWithUser verifica se esiste già una chat privata (non gruppo) tra due utenti specifici.
// Restituisce una lista contenente la chat trovata (se esiste), completa di informazioni sull'altro utente.
func (db *appdbimpl) GetChatWithUser(userId int, targetUsername string) ([]ChatListItem, error) {

	var chats []ChatListItem

	var targetId int
	var targetPhoto sql.NullString
	err := db.c.QueryRow("SELECT id, profile_photo FROM users WHERE username = ?", targetUsername).Scan(&targetId, &targetPhoto)
	if err != nil {
		return chats, nil
	}

	query := `
		SELECT c.id, c.is_group
		FROM chats c
		JOIN members m1 ON c.id = m1.chat_id
		JOIN members m2 ON c.id = m2.chat_id
		WHERE m1.user_id = ? AND m2.user_id = ? AND c.is_group = FALSE
	`
	rows, err := db.c.Query(query, userId, targetId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c ChatListItem
		c.SnippetText = ""
		if err := rows.Scan(&c.Id, &c.IsGroup); err != nil {
			return nil, err
		}

		c.Name = targetUsername
		if targetPhoto.Valid {
			c.PhotoChat = targetPhoto.String
		} else {
			c.PhotoChat = ""
		}

		chats = append(chats, c)
	}

	return chats, nil
}

// CreateConversation inizializza una nuova conversazione privata tra due utenti.
// Crea una nuova voce nella tabella chats e associa entrambi gli utenti nella tabella members.
// Restituisce l'oggetto ChatListItem inizializzato per l'interfaccia utente.
func (db *appdbimpl) CreateConversation(user1 int, user2 int) (ChatListItem, error) {
	var chat ChatListItem

	res, err := db.c.Exec("INSERT INTO chats (is_group) VALUES (FALSE)")
	if err != nil {
		return chat, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return chat, err
	}
	chatId := int(lastId)

	_, err = db.c.Exec("INSERT INTO members (chat_id, user_id) VALUES (?, ?)", chatId, user1)
	if err != nil {
		return chat, err
	}

	_, err = db.c.Exec("INSERT INTO members (chat_id, user_id) VALUES (?, ?)", chatId, user2)
	if err != nil {
		return chat, err
	}

	var otherName string
	var otherPhoto sql.NullString
	err = db.c.QueryRow("SELECT username, profile_photo FROM users WHERE id = ?", user2).Scan(&otherName, &otherPhoto)
	if err != nil {
		otherName = "Utente"
	}

	chat.Id = chatId
	chat.IsGroup = false
	chat.SnippetText = "Nuova chat"
	chat.Name = otherName

	if otherPhoto.Valid {
		chat.PhotoChat = otherPhoto.String
	} else {
		chat.PhotoChat = ""
	}

	return chat, nil
}

// GetConversation recupera la cronologia completa dei messaggi di una chat specifica.
// Esegue una join con la tabella utenti per includere il nome del mittente in ogni messaggio.
// Inoltre, marca automaticamente come letti tutti i messaggi inviati dagli altri partecipanti.
func (db *appdbimpl) GetConversation(userId int, chatId int) ([]Message, error) {

	_, err := db.c.Exec("UPDATE messages SET is_read = TRUE WHERE chat_id = ? AND user_id != ?", chatId, userId)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT 
			m.id, m.chat_id, m.user_id, m.text, COALESCE(m.photo_url, ''), m.sent_at, m.is_read,
			COALESCE(m.reply_to_message_id, 0), m.is_forward,
			u.username 
		FROM messages m
		JOIN users u ON m.user_id = u.id
		WHERE m.chat_id = ?
		ORDER BY m.sent_at ASC
	`
	rows, err := db.c.Query(query, chatId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]Message, 0)
	for rows.Next() {

		var m Message

		if err := rows.Scan(&m.Id, &m.ChatId, &m.SentBy, &m.Text, &m.PhotoUrl, &m.SentAt, &m.Checkmark, &m.ReplyTo, &m.IsForward, &m.SenderName); err != nil {
			continue
		}

		m.Reactions = make([]Reaction, 0)
		reactionRows, err := db.c.Query("SELECT r.emoticon, u.username FROM reactions r JOIN users u ON r.user_id = u.id WHERE r.message_id = ?", m.Id)
		if err == nil {
			for reactionRows.Next() {
				var r Reaction
				if err := reactionRows.Scan(&r.Emoticon, &r.Username); err == nil {
					m.Reactions = append(m.Reactions, r)
				}
			}
			reactionRows.Close()
		}

		messages = append(messages, m)
	}
	return messages, nil
}
