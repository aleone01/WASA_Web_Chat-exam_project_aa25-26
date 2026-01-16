package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// User rappresenta la struttura dati core di un utente nel sistema WASAPhoto.
// Mappa direttamente le colonne della tabella 'users'.
type User struct {
	Id           int    `json:"id"`
	Username     string `json:"username"`
	ProfilePhoto []byte `json:"profilePhoto"` // Trasmesso come stringa base64 nel JSON se gestito automaticamente
}

// ChatListItem è un DTO (Data Transfer Object) ottimizzato per la lista delle conversazioni (Home Screen).
// Aggrega dati provenienti da più tabelle (chats, messages, users) per mostrare un'anteprima.
type ChatListItem struct {
	Id          int       `json:"id"`
	Name        string    `json:"name,omitempty"` // Nome del gruppo o dell'interlocutore
	IsGroup     bool      `json:"isGroup"`
	PhotoChat   []byte    `json:"photoChat"`
	LastMessage time.Time `json:"lastMessage"` // Timestamp per ordinamento
	SnippetText string    `json:"snippetText"` // Anteprima testuale dell'ultimo messaggio
	SnippetIcon string    `json:"snippetIcon"` // Eventuale icona di stato
}

// Message rappresenta un singolo messaggio scambiato.
// Include metadati per funzionalità avanzate come inoltro, risposte e stato di lettura (spunte).
type Message struct {
	Id         int        `json:"id"`
	ChatId     int        `json:"-"` // Non esposto nel JSON, uso interno
	Text       string     `json:"text"`
	SentAt     time.Time  `json:"sentAt"`
	SentBy     int        `json:"sentBy"`
	PhotoFile  []byte     `json:"photoFile,omitempty"`
	Checkmark  bool       `json:"checkmark"`  // true se letto, false se inviato/consegnato
	ReplyTo    int        `json:"replyTo"`    // ID del messaggio a cui questo risponde (0 se nessuno)
	IsForward  bool       `json:"isForward"`  // true se il messaggio è stato inoltrato da un'altra chat
	SenderName string     `json:"senderName"` // Nome visualizzato del mittente (join con users)
	Reactions  []Reaction `json:"reactions"`  // Lista delle reazioni associate
}

// Reaction modella l'associazione tra un messaggio e un'emoticon aggiunta da un utente.
type Reaction struct {
	Emoticon string `json:"emoticon"`
	Username string `json:"username"` // Chi ha messo la reazione
}

// Group contiene i dettagli strutturali di una chat di gruppo.
// Utilizzato principalmente durante la creazione o la modifica delle impostazioni del gruppo.
type Group struct {
	Id          int    `json:"id"`
	GroupName   string `json:"groupname"`
	GroupPhoto  []byte `json:"groupPhoto"`
	MembersList []int  `json:"membersList"` // Lista degli ID degli utenti partecipanti
}

// AppDatabase definisce il contratto per il layer di persistenza dati.
// Questa interfaccia permette di disaccoppiare la logica API dall'implementazione specifica (es. SQLite),
// facilitando i test (tramite mock) e la manutenibilità.
type AppDatabase interface {
	Ping() error

	// Gestione Utenti
	CheckToken(token string) (int, error)
	UserLogin(username string) (int, bool, error)
	GetUserByUsername(username string) (User, error)
	SetMyUsername(userId int, newUsername string) (User, error)
	SetProfilePhoto(userId int, photoFile []byte) (User, error)

	// Gestione Chat e Conversazioni
	CreateConversation(user1 int, user2 int) (ChatListItem, error)
	GetMyConversations(userId int) ([]ChatListItem, error)
	GetConversation(userId int, chatId int) ([]Message, error)
	GetChatWithUser(userId int, username string) ([]ChatListItem, error)

	// Gestione Messaggi
	CreateMessage(chatId int, userId int, text string, photoFile []byte, sentAt time.Time, replyTo int, isForward bool) (Message, error)
	DeleteMessage(messageId int, userId int) error
	GetMessage(messageId int) (Message, error)

	// Gestione Reazioni
	AddReaction(messageId int, userId int, emoticon string) error
	RemoveReaction(messageId int, userId int) error

	// Gestione Gruppi
	CreateGroup(name string, photo []byte, members []int, creatorId int) (Group, error)
	AddGroupMembers(groupId int, newMembers []int) (Group, error)
	LeaveGroup(groupId int, userId int) error
	SetGroupName(groupId int, newName string) (Group, error)
	SetGroupPhoto(groupId int, photoFile []byte) (Group, error)
	CheckGroupMembership(groupId int, userId int) (bool, error)
	GetGroupById(groupId int) (Group, error)
	GetGroupMembers(groupId int) ([]User, error)
}

// appdbimpl è l'implementazione concreta di AppDatabase per SQLite.
type appdbimpl struct {
	c *sql.DB
}

// New inizializza il database e lo schema relazionale.
// 1. Abilita le foreign keys (disabilitate di default in SQLite).
// 2. Crea le tabelle se non esistono:
//   - users: anagrafica utenti e token.
//   - chats: registro conversazioni (gruppi e private).
//   - members: tabella di associazione users <-> chats (relazione molti-a-molti).
//   - messages: storico messaggi con foreign key verso chat e utenti.
//   - reactions: reazioni ai messaggi.
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}

	// Abilitazione vincoli di integrità referenziale per CASCADE DELETE
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("error enabling foreign keys: %w", err)
	}

	var schema = `
	CREATE TABLE IF NOT EXISTS users 
	(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		profile_photo BLOB,
		token TEXT
	);

	CREATE TABLE IF NOT EXISTS chats 
	(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		is_group BOOLEAN DEFAULT FALSE,
		group_name TEXT,
		group_photo BLOB
	);

	CREATE TABLE IF NOT EXISTS members 
	(
		chat_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		PRIMARY KEY (chat_id, user_id),
		FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS messages 
	(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		text TEXT,
		photo_file BLOB,
		sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		is_read BOOLEAN DEFAULT FALSE,
		reply_to_message_id INTEGER,
		is_forward BOOLEAN DEFAULT FALSE,
		FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (reply_to_message_id) REFERENCES messages(id) ON DELETE SET NULL
	);

	CREATE TABLE IF NOT EXISTS reactions 
	(
		message_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		emoticon TEXT NOT NULL,
		PRIMARY KEY (message_id, user_id),
		FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	`
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("error creating database structure: %w", err)
	}

	return &appdbimpl{
		c: db,
	}, nil
}

// Ping controlla che la connessione al database sia attiva e funzionante.
func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}
