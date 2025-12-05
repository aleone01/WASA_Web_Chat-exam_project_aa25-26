package database

import (
	"database/sql"
	"errors"
)

// CreateGroup crea un nuovo gruppo e aggiunge i membri iniziali (incluso il creatore)
func (db *appdbimpl) CreateGroup(name string, photo string, members []int, creatorId int) (Group, error) {
	var g Group

	// inserimento nella tabella chats
	res, err := db.c.Exec("INSERT INTO chats (is_group, group_name, group_photo) VALUES (TRUE, ?, ?)", name, photo)
	if err != nil {
		return g, err
	}

	// recupero ID del gruppo appena creato
	groupId, err := res.LastInsertId()
	if err != nil {
		return g, err
	}

	// aggiunge il creatore come membro
	_, err = db.c.Exec("INSERT INTO members (chat_id, user_id) VALUES (?, ?)", groupId, creatorId)
	if err != nil {
		return g, err
	}

	// aggiunge gli altri membri
	for _, memberId := range members {
		// evita di riaggiungere il creatore
		if memberId == creatorId {
			continue
		}
		_, err = db.c.Exec("INSERT INTO members (chat_id, user_id) VALUES (?, ?)", groupId, memberId)
		// ignora errori di inserimento
		if err != nil {
			continue
		}
	}

	// recupera e ritorna il gruppo creato
	return db.GetGroupById(int(groupId))

}

// AddGroupMembers aggiunge una lista di utenti a un gruppo esistente
func (db *appdbimpl) AddGroupMembers(groupId int, newMembers []int) (Group, error) {

	// aggiunge ciascun nuovo membro
	for _, memberId := range newMembers {
		_, err := db.c.Exec("INSERT INTO members (chat_id, user_id) VALUES (?, ?)", groupId, memberId)
		if err != nil {
			continue
		}
	}

	// recupera e ritorna il gruppo aggiornato
	return db.GetGroupById(groupId)
}

// LeaveGroup rimuove un utente (se stesso) dal gruppo
func (db *appdbimpl) LeaveGroup(groupId int, userId int) error {

	// esecuzione DELETE
	res, err := db.c.Exec("DELETE FROM members WHERE chat_id = ? AND user_id = ?", groupId, userId)
	if err != nil {
		return err
	}

	// controlla se è stato effettivamente rimosso
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("group not found or user not member")
	}
	return nil
}

// SetGroupName aggiorna il nome del gruppo
func (db *appdbimpl) SetGroupName(groupId int, newName string) (Group, error) {

	// esecuzione UPDATE
	_, err := db.c.Exec("UPDATE chats SET group_name = ? WHERE id = ? AND is_group = TRUE", newName, groupId)
	if err != nil {
		return Group{}, err
	}

	// recupera e ritorna il gruppo aggiornato
	return db.GetGroupById(groupId)
}

// SetGroupPhoto aggiorna la foto del gruppo
func (db *appdbimpl) SetGroupPhoto(groupId int, photoURL string) (Group, error) {

	// esecuzione UPDATE
	_, err := db.c.Exec("UPDATE chats SET group_photo = ? WHERE id = ? AND is_group = TRUE", photoURL, groupId)
	if err != nil {
		return Group{}, err
	}

	// recupera e ritorna il gruppo aggiornato
	return db.GetGroupById(groupId)
}

// CheckGroupMembership verifica se un utente fa parte di un gruppo
func (db *appdbimpl) CheckGroupMembership(groupId int, userId int) (bool, error) {

	// esecuzione SELECT
	var exists int
	err := db.c.QueryRow("SELECT 1 FROM members WHERE chat_id = ? AND user_id = ?", groupId, userId).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return true, err
}

// GetGroupById recupera tutti i dati di un gruppo, inclusa la lista membri
func (db *appdbimpl) GetGroupById(groupId int) (Group, error) {

	// crea oggetto gruppo
	var g Group
	g.Id = groupId

	// recupero dati base del gruppo
	err := db.c.QueryRow("SELECT group_name, COALESCE(group_photo, '') FROM chats WHERE id = ? AND is_group = TRUE", groupId).Scan(&g.GroupName, &g.GroupPhoto)
	if err != nil {
		return g, err
	}

	// recupero lista membri
	rows, err := db.c.Query("SELECT user_id FROM members WHERE chat_id = ?", groupId)
	if err != nil {
		return g, err
	}
	defer rows.Close()

	// ciclo sui risultati
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

	// ritorna il gruppo completo
	return g, nil
}
