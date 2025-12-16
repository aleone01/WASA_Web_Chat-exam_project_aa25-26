package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// User rappresenta la struttura dati di un utente registrato nel sistema.
// Contiene l'identificativo univoco, il nome utente e l'URL (opzionale) della foto profilo.
type User struct {
	Id           int    `json:"id"`
	Username     string `json:"username"`
	ProfilePhoto string `json:"profilePhoto"`
}

// ChatListItem definisce il modello per visualizzare l'anteprima di una conversazione nella lista delle chat.
// Include informazioni come il nome della chat (o del gruppo), l'ultimo messaggio inviato e lo stato (gruppo o chat privata).
type ChatListItem struct {
	Id          int       `json:"id"`
	Name        string    `json:"name,omitempty"`
	IsGroup     bool      `json:"isGroup"`
	PhotoChat   string    `json:"photoChat"`
	LastMessage time.Time `json:"lastMessage"`
	SnippetText string    `json:"snippetText"`
	SnippetIcon string    `json:"snippetIcon"`
}

// Message modella un singolo messaggio scambiato all'interno di una conversazione.
// Oltre al contenuto (testo/foto), tiene traccia del mittente, del timestamp, dello stato di lettura,
// dei riferimenti (reply/forward) e include il nome del mittente per facilitare la visualizzazione nel frontend.
type Message struct {
	Id         int        `json:"id"`
	ChatId     int        `json:"-"`
	Text       string     `json:"text"`
	SentAt     time.Time  `json:"sentAt"`
	SentBy     int        `json:"sentBy"`
	PhotoUrl   string     `json:"photoUrl,omitempty"`
	Checkmark  bool       `json:"checkmark"`
	ReplyTo    int        `json:"replyTo"`
	IsForward  bool       `json:"isForward"`
	SenderName string     `json:"senderName"`
	Reactions  []Reaction `json:"reactions"`
}

// Reaction rappresenta una singola reazione (emoticon) lasciata da un utente su un messaggio.
// Include l'emoticon stessa e lo username dell'autore, permettendo al frontend di mostrare "chi ha reagito".
type Reaction struct {
	Emoticon string `json:"emoticon"`
	Username string `json:"username"`
}

// Group rappresenta i dettagli di un gruppo di conversazione.
// Contiene i metadati del gruppo (nome, foto) e la lista degli ID degli utenti partecipanti.
type Group struct {
	Id          int    `json:"id"`
	GroupName   string `json:"groupname"`
	GroupPhoto  string `json:"groupPhoto"`
	MembersList []int  `json:"membersList"`
}

// AppDatabase è l'interfaccia principale che astrae tutte le interazioni con il database.
// Definisce i metodi necessari per la gestione di utenti, autenticazione, chat (private e gruppi),
// messaggistica e reazioni, disaccoppiando la logica di business dall'implementazione SQL sottostante.
type AppDatabase interface {
	Ping() error
	CheckToken(token string) (int, error)
	UserLogin(username string) (int, bool, error)
	GetUserByUsername(username string) (User, error)
	SetMyUsername(userId int, newUsername string) (User, error)
	SetProfilePhoto(userId int, photoURL string) (User, error)
	CreateConversation(user1 int, user2 int) (ChatListItem, error)
	GetMyConversations(userId int) ([]ChatListItem, error)
	GetConversation(userId int, chatId int) ([]Message, error)
	GetChatWithUser(userId int, username string) ([]ChatListItem, error)
	CreateMessage(chatId int, userId int, text string, photoUrl string, sentAt time.Time, replyTo int, isForward bool) (Message, error)
	AddReaction(messageId int, userId int, emoticon string) error
	RemoveReaction(messageId int, userId int) error
	DeleteMessage(messageId int, userId int) error
	GetMessage(messageId int) (Message, error)
	CreateGroup(name string, photo string, members []int, creatorId int) (Group, error)
	AddGroupMembers(groupId int, newMembers []int) (Group, error)
	LeaveGroup(groupId int, userId int) error
	SetGroupName(groupId int, newName string) (Group, error)
	SetGroupPhoto(groupId int, photoURL string) (Group, error)
	CheckGroupMembership(groupId int, userId int) (bool, error)
	GetGroupById(groupId int) (Group, error)
	GetGroupMembers(groupId int) ([]User, error)
}

// appdbimpl è l'implementazione concreta dell'interfaccia AppDatabase basata su database/sql (SQLite).
type appdbimpl struct {
	c *sql.DB
}

// New inizializza una nuova istanza del database.
// Configura le opzioni di connessione (es. chiavi esterne per SQLite) e crea lo schema delle tabelle
// (utenti, chat, membri, messaggi, reazioni) se non esistono già.
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("error enabling foreign keys: %w", err)
	}

	var schema = `
	CREATE TABLE IF NOT EXISTS users 
	(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		profile_photo TEXT,
		token TEXT
	);

	CREATE TABLE IF NOT EXISTS chats 
	(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		is_group BOOLEAN DEFAULT FALSE,
		group_name TEXT,
		group_photo TEXT
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
		photo_url TEXT,
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

// Ping verifica la connettività al database sottostante.
func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}
