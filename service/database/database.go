package database

import 
(
	"database/sql"
	"errors"
	"fmt"
)

// strutture dati usate nelle API
// User rappresenta un utente del sistema
type User struct 
{
	Id           int    `json:"id"`
	Username     string `json:"username"`
	ProfilePhoto string `json:"profilePhoto"`
}

// ChatListItem rappresenta un elemento della lista delle chat
type ChatListItem struct 
{
	Id          int       `json:"id"`
	IsGroup     bool      `json:"isGroup"`
	PhotoChat   string    `json:"photoChat"`
	LastMessage time.Time `json:"lastMessage"`
	SnippetText string    `json:"snippetText"`
	SnippetIcon string    `json:"snippetIcon"`
}

// Message rappresenta un messaggio in una chat
type Message struct 
{
	Id        int       `json:"id"`
	ChatId    int       `json:"-"` 
	Text      string    `json:"text"`
	SentAt    time.Time `json:"sentAt"`
	SentBy    int       `json:"sentBy"`
	PhotoUrl  string    `json:"photoUrl,omitempty"`
	Checkmark bool      `json:"checkmark"`
}

// Group rappresenta un gruppo di chat
type Group struct 
{
	Id          int    `json:"id"`
	GroupName   string `json:"groupname"`
	GroupPhoto  string `json:"groupPhoto"`
	MembersList []int  `json:"membersList"`
}


// AppDatabase is the high level interface for the DB
// Qui definiamo i metodi che il resto dell'applicazione potrà chiamare.
type AppDatabase interface 
{
	Ping() error

	// CheckToken verifica se il token di sessione è valido e restituisce l'userId associato
	CheckToken(token string) (int, error)

	// UserLogin verifica se l'utente esiste: se sì restituisce il suo ID e 'false', se no lo crea e restituisce il suo ID e 'true'.
	UserLogin(username string) (int, bool, error)

	// SetUsername aggiorna lo username dell'utente, riconosciuto con userId
	SetMyUsername(userId int, newUsername string) (User, error)

	// SetProfilePhoto aggiorna la foto profilo dell'utente, riconosciuto con userId
	SetProfilePhoto(userId int, photoURL string) (User, error)

	// GetConversations restituisce la lista delle conversazioni dell'utente ordinate cronologicamente
	GetConversations(userId int) ([]ChatListItem, error)

	// GetChatWithUser cerca le conversazioni dell'utente con un altro utente specifico (username)
	GetChatWithUser(userId int, username string) ([]ChatListItem, error)

	// GetChatMessages restituisce i messaggi di una chat specifica, se l'userId è partecipante della chat
	GetChatMessages(chatId int, userId int) ([]Message, error)

	// CreateMessage crea un nuovo messaggio in una chat
	CreateMessage(chatId int, userId int, text string, photoUrl string) (Message, error)

	// AddReaction aggiunge un'emoticon a un messaggio
	AddReaction(messageId int, userId int, emoticon string) error

	// RemoveReaction rimuove la reazione dell'utente da un messaggio 
	RemoveReaction(messageId int, userId int) error

	// DeleteMessage elimina un messaggio (solo se l'utente ne è il proprietario)
	DeleteMessage(messageId int, userId int) error

	// GetMessage recupera un messaggio dal suo ID 
	GetMessage(messageId int) (Message, error)

	// CreateGroup crea un nuovo gruppo con i membri specificati
	CreateGroup(name string, photo string, members []int, creatorId int) (Group, error)

	// AddGroupMembers aggiunge nuovi membri a un gruppo esistente
    AddGroupMembers(groupId int, newMembers []int) (Group, error)

	// LeaveGroup rimuove un utente (se stesso) dal gruppo
    LeaveGroup(groupId int, userId int) error

	// SetGroupName aggiorna il nome del gruppo
    SetGroupName(groupId int, newName string) (Group, error)

	// SetGroupPhoto aggiorna la foto del gruppo
    SetGroupPhoto(groupId int, photoURL string) (Group, error)

	// CheckGroupMembership verifica se un utente è membro di un gruppo
    CheckGroupMembership(groupId int, userId int) (bool, error)

	// GetGroupById recupera un gruppo dal suo ID
    GetGroupById(groupId int) (Group, error)

}

// appdbimpl è l'implementazione concreta di AppDatabase
type appdbimpl struct 
{
	c *sql.DB
}

// New returns a new instance of AppDatabase based on the SQLite connection `db`.
// `db` is required - an error will be returned if `db` is `nil`.
func New(db *sql.DB) (AppDatabase, error) 
{
	// controllo che il db non sia nil
	if db == nil 
	{
		return nil, errors.New("database is required when building a AppDatabase")
	}

	// abilita le Foreign Keys 
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil 
	{
		return nil, fmt.Errorf("error enabling foreign keys: %w", err)
	}

	// tabelle sqlite create sulla base degli schemas di api.yaml
	var schema = `

	// tabella utenti
	CREATE TABLE IF NOT EXISTS users 
	(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		profile_photo TEXT
	);

	// tabella chat
	CREATE TABLE IF NOT EXISTS chats 
	(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		is_group BOOLEAN DEFAULT FALSE,
		group_name TEXT,
		group_photo TEXT
	);

	// tabella membri delle chat
	CREATE TABLE IF NOT EXISTS members 
	(
		chat_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		PRIMARY KEY (chat_id, user_id),
		FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	// tabella messaggi
	CREATE TABLE IF NOT EXISTS messages 
	(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		text TEXT,
		photo_url TEXT,
		sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		reply_to_message_id INTEGER,
		FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (reply_to_message_id) REFERENCES messages(id) ON DELETE SET NULL
	);

	// tabella reazioni ai messaggi
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

	// esecuzione creazione tabelle
	if _, err := db.Exec(schema); err != nil 
	{
		return nil, fmt.Errorf("error creating database structure: %w", err)
	}

	// restituisce l'istanza di appdbimpl
	return &appdbimpl
	{
		c: db,
	}, nil
}

// Ping verifica la connessione al database
func (db *appdbimpl) Ping() error 
{
	return db.c.Ping()
}