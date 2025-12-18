package database

import (
	"database/sql"
	"errors"
)

// CreateGroup crea un nuovo gruppo di conversazione.
// Inserisce i metadati del gruppo (nome, foto) nella tabella chats e popola la tabella members
// associando il creatore e la lista degli altri membri iniziali al nuovo gruppo.
func (db *appdbimpl) CreateGroup(name string, photo string, members []int, creatorId int) (Group, error) {
	var g Group

	res, err := db.c.Exec("INSERT INTO chats (is_group, group_name, group_photo) VALUES (TRUE, ?, ?)", name, photo)
	if err != nil {
		return g, err
	}

	groupId, err := res.LastInsertId()
	if err != nil {
		return g, err
	}

	_, err = db.c.Exec("INSERT INTO members (chat_id, user_id) VALUES (?, ?)", groupId, creatorId)
	if err != nil {
		return g, err
	}

	for _, memberId := range members {
		if memberId == creatorId {
			continue
		}
		_, err = db.c.Exec("INSERT INTO members (chat_id, user_id) VALUES (?, ?)", groupId, memberId)
		if err != nil {
			continue
		}
	}

	return db.GetGroupById(int(groupId))

}

// AddGroupMembers aggiunge una lista di nuovi utenti a un gruppo esistente.
// Itera sulla lista di ID fornita e inserisce le nuove associazioni nella tabella members, ignorando eventuali duplicati o errori.
func (db *appdbimpl) AddGroupMembers(groupId int, newMembers []int) (Group, error) {

	for _, memberId := range newMembers {
		_, err := db.c.Exec("INSERT INTO members (chat_id, user_id) VALUES (?, ?)", groupId, memberId)
		if err != nil {
			continue
		}
	}

	return db.GetGroupById(groupId)
}

// LeaveGroup rimuove l'associazione tra un utente e un gruppo specifico.
// Se l'operazione non elimina alcuna riga (es. utente non membro o gruppo inesistente), restituisce un errore.
func (db *appdbimpl) LeaveGroup(groupId int, userId int) error {

	res, err := db.c.Exec("DELETE FROM members WHERE chat_id = ? AND user_id = ?", groupId, userId)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("group not found or user not member")
	}
	return nil
}

// SetGroupName modifica il nome visualizzato di un gruppo.
// Esegue un aggiornamento sulla tabella chats e restituisce l'oggetto Group aggiornato.
func (db *appdbimpl) SetGroupName(groupId int, newName string) (Group, error) {

	_, err := db.c.Exec("UPDATE chats SET group_name = ? WHERE id = ? AND is_group = TRUE", newName, groupId)
	if err != nil {
		return Group{}, err
	}

	return db.GetGroupById(groupId)
}

// SetGroupPhoto aggiorna l'immagine rappresentativa di un gruppo.
// Aggiorna il campo group_photo nel database e restituisce i dati aggiornati del gruppo.
func (db *appdbimpl) SetGroupPhoto(groupId int, photoURL string) (Group, error) {

	_, err := db.c.Exec("UPDATE chats SET group_photo = ? WHERE id = ? AND is_group = TRUE", photoURL, groupId)
	if err != nil {
		return Group{}, err
	}

	return db.GetGroupById(groupId)
}

// CheckGroupMembership verifica se un dato utente è un membro attivo di un gruppo specifico.
// Restituisce true se l'associazione esiste nella tabella members, false altrimenti.
func (db *appdbimpl) CheckGroupMembership(groupId int, userId int) (bool, error) {

	var exists int
	err := db.c.QueryRow("SELECT 1 FROM members WHERE chat_id = ? AND user_id = ?", groupId, userId).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return true, err
}

// GetGroupById recupera le informazioni complete di un gruppo dato il suo ID.
// Esegue due query: una per i metadati del gruppo (nome, foto) e una per ottenere la lista degli ID di tutti i membri.
func (db *appdbimpl) GetGroupById(groupId int) (Group, error) {

	var g Group
	g.Id = groupId

	err := db.c.QueryRow("SELECT group_name, COALESCE(group_photo, '') FROM chats WHERE id = ? AND is_group = TRUE", groupId).Scan(&g.GroupName, &g.GroupPhoto)
	if err != nil {
		return g, err
	}

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

// GetGroupMembers recupera la lista completa degli utenti che fanno parte di un gruppo specifico.
// Esegue una JOIN tra la tabella 'members' e la tabella 'users' per ottenere username e foto.
// GetGroupMembers recupera la lista completa degli utenti che fanno parte di un gruppo specifico.
func (db *appdbimpl) GetGroupMembers(groupId int) ([]User, error) {

	users := make([]User, 0)

	query := `
		SELECT u.id, u.username, COALESCE(u.profile_photo, '')
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
			users = append(users, u)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
