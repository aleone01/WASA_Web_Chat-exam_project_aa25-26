package database

import (
	"database/sql"
)

// GetMyConversations è la funzione principale per popolare la schermata "Chat" dell'app.
// Recupera tutte le chat (gruppi e private) associate all'utente.
// Per ogni chat calcola dinamicamente:
// - L'anteprima dell'ultimo messaggio (testo o "📷 Foto").
// - Il timestamp dell'ultimo messaggio per l'ordinamento.
// - Il nome da visualizzare: nome del gruppo (se gruppo) o username dell'altro utente (se chat privata).
// - La foto da visualizzare: foto del gruppo o foto profilo dell'altro utente.
func (db *appdbimpl) GetMyConversations(userId int) ([]ChatListItem, error) {

	chats := make([]ChatListItem, 0)

	// Query complessa che usa JOIN multiple e subquery correlate per estrarre l'ultimo messaggio per ogni chat.
	query := `
		SELECT 
			c.id, 
			c.is_group, 
			COALESCE(c.group_name, ''),
			c.group_photo,
			(SELECT sent_at FROM messages WHERE chat_id = c.id ORDER BY sent_at DESC LIMIT 1) as last_msg_time,
			(SELECT text FROM messages WHERE chat_id = c.id ORDER BY sent_at DESC LIMIT 1) as snippet_text,
			(SELECT photo_file FROM messages WHERE chat_id = c.id ORDER BY sent_at DESC LIMIT 1) as snippet_photo,
			COALESCE(u.username, ''),
			u.profile_photo
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

	// Parsing dei risultati
	for rows.Next() {
		var c ChatListItem
		var groupName string
		var groupPhoto []byte
		var lastTime sql.NullTime
		var lastText sql.NullString
		var lastPhoto []byte
		var otherUsername sql.NullString
		var otherPhoto []byte

		c.SnippetIcon = ""

		// Parsing dei risultati che possono contenere valori NULL (sql.Null*)
		if err := rows.Scan(&c.Id, &c.IsGroup, &groupName, &groupPhoto, &lastTime, &lastText, &lastPhoto, &otherUsername, &otherPhoto); err != nil {
			continue
		}

		if lastTime.Valid {
			c.LastMessage = lastTime.Time
		}

		// Logica per determinare lo snippet (anteprima)
		if lastText.Valid && lastText.String != "" {
			c.SnippetText = lastText.String
		} else if len(lastPhoto) > 0 {
			c.SnippetText = "📷 Foto"
		} else {
			c.SnippetText = ""
		}

		// Logica per determinare Nome e Icona della chat
		if c.IsGroup {
			c.Name = groupName
			c.PhotoChat = groupPhoto
		} else {
			// Se è una chat privata, usiamo i dati dell'interlocutore
			if otherUsername.Valid {
				c.Name = otherUsername.String
			} else {
				c.Name = "Utente Sconosciuto"
			}
			c.PhotoChat = otherPhoto
		}

		if c.PhotoChat == nil {
			c.PhotoChat = []byte{}
		}

		chats = append(chats, c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return chats, nil
}

// GetChatWithUser cerca se esiste una conversazione diretta (1-a-1) tra l'utente richiedente e un target.
// Viene usata per evitare di creare chat duplicate quando si clicca "Nuova Chat" con un utente con cui si è già parlato.
func (db *appdbimpl) GetChatWithUser(userId int, targetUsername string) ([]ChatListItem, error) {

	var chats []ChatListItem

	// Recupero ID dell'utente target
	var targetId int
	var targetPhoto []byte
	err := db.c.QueryRow("SELECT id, profile_photo FROM users WHERE username = ?", targetUsername).Scan(&targetId, &targetPhoto)
	if err != nil {
		return chats, nil
	}

	// Query che cerca una chat non-gruppo che contenga entrambi gli ID utente
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

	// Parsing dei risultati
	for rows.Next() {
		var c ChatListItem
		c.SnippetText = ""
		if err := rows.Scan(&c.Id, &c.IsGroup); err != nil {
			return nil, err
		}

		c.Name = targetUsername
		c.PhotoChat = targetPhoto
		if c.PhotoChat == nil {
			c.PhotoChat = []byte{}
		}

		chats = append(chats, c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return chats, nil
}

// CreateConversation crea una nuova chat privata (Direct Message) tra due utenti.
// Inserisce una riga in 'chats' e due righe in 'members'.
// Restituisce l'oggetto ChatListItem pronto per essere aggiunto alla lista chat del frontend.
func (db *appdbimpl) CreateConversation(user1 int, user2 int) (ChatListItem, error) {
	var chat ChatListItem

	// Creazione chat
	res, err := db.c.Exec("INSERT INTO chats (is_group) VALUES (FALSE)")
	if err != nil {
		return chat, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return chat, err
	}
	chatId := int(lastId)

	// Associazione Membro 1
	_, err = db.c.Exec("INSERT INTO members (chat_id, user_id) VALUES (?, ?)", chatId, user1)
	if err != nil {
		return chat, err
	}

	// Associazione Membro 2
	_, err = db.c.Exec("INSERT INTO members (chat_id, user_id) VALUES (?, ?)", chatId, user2)
	if err != nil {
		return chat, err
	}

	// Recupero info secondo utente per popolare il nome chat
	var otherName string
	var otherPhoto []byte
	err = db.c.QueryRow("SELECT username, profile_photo FROM users WHERE id = ?", user2).Scan(&otherName, &otherPhoto)
	if err != nil {
		otherName = "Utente"
	}

	chat.Id = chatId
	chat.IsGroup = false
	chat.SnippetText = "Nuova chat"
	chat.Name = otherName
	chat.PhotoChat = otherPhoto
	if chat.PhotoChat == nil {
		chat.PhotoChat = []byte{}
	}

	return chat, nil
}

// GetConversation recupera l'intera history dei messaggi di una chat.
// Esegue due operazioni importanti:
// 1. "Side Effect": Segna come LETTI (is_read=TRUE) tutti i messaggi inviati dagli altri utenti in questa chat (Read Receipt).
// 2. Recupero dati: Estrae messaggi ordinati per data e arricchiti con reazioni.
func (db *appdbimpl) GetConversation(userId int, chatId int) ([]Message, error) {

	// Update per le spunte blu (letto)
	_, err := db.c.Exec("UPDATE messages SET is_read = TRUE WHERE chat_id = ? AND user_id != ?", chatId, userId)
	if err != nil {
		return nil, err
	}

	// Query principale messaggi + JOIN username mittente
	query := `
		SELECT 
			m.id, m.chat_id, m.user_id, m.text, m.photo_file, m.sent_at, m.is_read,
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

		if err := rows.Scan(&m.Id, &m.ChatId, &m.SentBy, &m.Text, &m.PhotoFile, &m.SentAt, &m.Checkmark, &m.ReplyTo, &m.IsForward, &m.SenderName); err != nil {
			continue
		}

		if m.PhotoFile == nil {
			m.PhotoFile = []byte{}
		}

		// Per ogni messaggio, recuperiamo le eventuali reazioni
		m.Reactions = make([]Reaction, 0)
		reactionRows, err := db.c.Query("SELECT r.emoticon, u.username FROM reactions r JOIN users u ON r.user_id = u.id WHERE r.message_id = ?", m.Id)
		if err == nil {
			for reactionRows.Next() {

				var r Reaction
				if err := reactionRows.Scan(&r.Emoticon, &r.Username); err == nil {
					m.Reactions = append(m.Reactions, r)
				}

				if err := reactionRows.Err(); err != nil {
					reactionRows.Close()
					return nil, err
				}
			}
			reactionRows.Close()
		}

		messages = append(messages, m)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}
