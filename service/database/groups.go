package database

import (
	"database/sql"
	"errors"
)

// CreateGroup gestisce la creazione di una chat di gruppo.
// Esegue una serie di operazioni sequenziali:
// 1. Valida che tutti gli ID utente forniti esistano nel sistema.
// 2. Crea il record nella tabella 'chats' con flag is_group=TRUE.
// 3. Associa il creatore del gruppo alla chat.
// 4. Associa tutti gli altri membri alla chat nella tabella 'members'.
// Restituisce l'oggetto Group completo appena creato.
func (db *appdbimpl) CreateGroup(name string, photo []byte, members []int, creatorId int) (Group, error) {
	var g Group

	// Validazione preliminare dei membri
	for _, memberId := range members {
		var exists int
		err := db.c.QueryRow("SELECT id FROM users WHERE id = ?", memberId).Scan(&exists)
		if err != nil {
			return g, errors.New("Gruppo non creato: uno o più membri non esistono")
		}
	}

	// Inserimento metadati gruppo
	res, err := db.c.Exec("INSERT INTO chats (is_group, group_name, group_photo) VALUES (TRUE, ?, ?)", name, photo)
	if err != nil {
		return g, err
	}

	groupId, err := res.LastInsertId()
	if err != nil {
		return g, err
	}

	// Aggiunta automatica del creatore ai membri
	_, err = db.c.Exec("INSERT INTO members (chat_id, user_id) VALUES (?, ?)", groupId, creatorId)
	if err != nil {
		return g, err
	}

	// Aggiunta degli altri partecipanti
	for _, memberId := range members {
		if memberId == creatorId {
			continue // Evita duplicati se il creatore si è auto-incluso nella lista
		}
		// INSERT ignorando errori (es. duplicati) per non bloccare l'intero processo
		_, err = db.c.Exec("INSERT INTO members (chat_id, user_id) VALUES (?, ?)", groupId, memberId)
		if err != nil {
			continue
		}
	}

	return db.GetGroupById(int(groupId))
}

// AddGroupMembers espande la membership di un gruppo esistente.
// Itera sulla lista di nuovi ID e tenta l'inserimento nella tabella 'members'.
// Restituisce lo stato aggiornato del gruppo.
func (db *appdbimpl) AddGroupMembers(groupId int, newMembers []int) (Group, error) {

	for _, memberId := range newMembers {
		// Tenta l'inserimento. Se l'utente è già membro (violazione Primary Key), l'errore viene ignorato e si prosegue.
		_, err := db.c.Exec("INSERT INTO members (chat_id, user_id) VALUES (?, ?)", groupId, memberId)
		if err != nil {
			continue
		}
	}

	return db.GetGroupById(groupId)
}

// LeaveGroup gestisce l'uscita di un utente da un gruppo.
// Esegue una DELETE sulla tabella 'members'.
// Restituisce errore se il gruppo non esiste o se l'utente non ne faceva parte.
func (db *appdbimpl) LeaveGroup(groupId int, userId int) error {

	// Esecuzione della DELETE
	res, err := db.c.Exec("DELETE FROM members WHERE chat_id = ? AND user_id = ?", groupId, userId)
	if err != nil {
		return err
	}

	// Verifica che almeno una riga sia stata cancellata
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("group not found or user not member")
	}
	return nil
}

// SetGroupName aggiorna il nome di un gruppo.
// Include un controllo di sicurezza nella query (is_group = TRUE) per evitare di rinominare chat private per errore.
func (db *appdbimpl) SetGroupName(groupId int, newName string) (Group, error) {

	// Esecuzione dell'aggiornamento
	_, err := db.c.Exec("UPDATE chats SET group_name = ? WHERE id = ? AND is_group = TRUE", newName, groupId)
	if err != nil {
		return Group{}, err
	}

	return db.GetGroupById(groupId)
}

// SetGroupPhoto aggiorna l'icona/foto del gruppo.
// Anche qui il vincolo is_group = TRUE protegge l'integrità dei dati.
func (db *appdbimpl) SetGroupPhoto(groupId int, photoFile []byte) (Group, error) {

	// Aggiornamento con BLOB
	_, err := db.c.Exec("UPDATE chats SET group_photo = ? WHERE id = ? AND is_group = TRUE", photoFile, groupId)
	if err != nil {
		return Group{}, err
	}

	return db.GetGroupById(groupId)
}

// CheckGroupMembership è una funzione di utility per verificare i permessi.
// Controlla se esiste una riga nella tabella 'members' per la coppia (groupId, userId).
func (db *appdbimpl) CheckGroupMembership(groupId int, userId int) (bool, error) {

	var exists int
	err := db.c.QueryRow("SELECT 1 FROM members WHERE chat_id = ? AND user_id = ?", groupId, userId).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return true, err
}

// GetGroupById ricostruisce l'oggetto Group recuperando dati da più query.
// 1. Recupera nome e foto dalla tabella 'chats'.
// 2. Recupera la lista degli ID utenti dalla tabella 'members'.
func (db *appdbimpl) GetGroupById(groupId int) (Group, error) {

	var g Group
	g.Id = groupId

	// Query metadati
	err := db.c.QueryRow("SELECT group_name, group_photo FROM chats WHERE id = ? AND is_group = TRUE", groupId).Scan(&g.GroupName, &g.GroupPhoto)
	if err != nil {
		return g, err
	}

	if g.GroupPhoto == nil {
		g.GroupPhoto = []byte{}
	}

	// Query lista membri
	rows, err := db.c.Query("SELECT user_id FROM members WHERE chat_id = ?", groupId)
	if err != nil {
		return g, err
	}
	defer rows.Close()

	for rows.Next() {
		var uid int
		if err := rows.Scan(&uid); err != nil {
			return g, err
		}
		g.MembersList = append(g.MembersList, uid)
	}

	if err := rows.Err(); err != nil {
		return g, err
	}

	return g, nil
}

// GetGroupMembers fornisce i dettagli completi (User struct) di tutti i partecipanti.
// Esegue una JOIN SQL esplicita tra 'members' e 'users' per efficienza.
func (db *appdbimpl) GetGroupMembers(groupId int) ([]User, error) {

	users := make([]User, 0)

	query := `
		SELECT u.id, u.username, u.profile_photo
		FROM members m
		JOIN users u ON m.user_id = u.id
		WHERE m.chat_id = ?
	`
	rows, err := db.c.Query(query, groupId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var u User
		if err := rows.Scan(&u.Id, &u.Username, &u.ProfilePhoto); err == nil {
			if u.ProfilePhoto == nil {
				u.ProfilePhoto = []byte{}
			}
			users = append(users, u)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
