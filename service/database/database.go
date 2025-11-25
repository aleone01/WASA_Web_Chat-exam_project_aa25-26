package database

import 
(
	"database/sql"
	"errors"
	"fmt"
)

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

}

type appdbimpl struct 
{
	c *sql.DB
}

// New returns a new instance of AppDatabase based on the SQLite connection `db`.
// `db` is required - an error will be returned if `db` is `nil`.
func New(db *sql.DB) (AppDatabase, error) 
{
	if db == nil 
	{
		return nil, errors.New("database is required when building a AppDatabase")
	}

	// 1. Abilitiamo le Foreign Keys (su SQLite sono disabilitate di default)
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil 
	{
		return nil, fmt.Errorf("error enabling foreign keys: %w", err)
	}

	// tabelle sqlite create sulla base degli schemas di api.yaml
	var schema = `
	CREATE TABLE IF NOT EXISTS users 
	(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		profile_photo TEXT
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
		reply_to_message_id INTEGER,
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

	// esecuzione creazione tabelle
	if _, err := db.Exec(schema); err != nil 
	{
		return nil, fmt.Errorf("error creating database structure: %w", err)
	}

	return &appdbimpl
	{
		c: db,
	}, nil
}

func (db *appdbimpl) Ping() error 
{
	return db.c.Ping()
}