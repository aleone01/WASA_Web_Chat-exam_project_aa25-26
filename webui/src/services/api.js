import axios from "./axios";

// Modulo API Service.
// Questo oggetto esporta tutte le funzioni necessarie per interagire con il backend.
// Astrae le chiamate HTTP dirette, offrendo ai componenti Vue metodi semantici (es. doLogin, sendMessage)
// e gestendo la costruzione dei payload JSON richiesti dalle API REST in Go.
export default {
	
    // Effettua il login inviando lo username al server e restituisce la risposta contenente l'ID utente/token.
	async doLogin(username) {
		return axios.post("/login", { username });
	},

	// Invia una richiesta PUT per aggiornare lo username dell'utente specificato.
	async setMyUserName(userId, newUsername) {
		return axios.put(`/users/${userId}/username`, { username: newUsername });
	},
	
    // Invia una richiesta PUT per aggiornare l'URL della foto profilo dell'utente.
	async setMyPhoto(userId, photoUrl) {
		return axios.put(`/users/${userId}/photo`, { photo: photoUrl });
	},

	// Recupera la lista di tutte le conversazioni (chat private e gruppi) associate all'utente.
	async getMyConversations(userId) {
		return axios.get(`/users/${userId}/chats`);
	},

    // Ottiene la cronologia dei messaggi per una specifica chat.
	async getConversation(userId, chatId) {
		return axios.get(`/users/${userId}/chats/${chatId}`);
	},

    // Crea una nuova conversazione privata con un altro utente specificato tramite username.
	async createConversation(userId, targetUsername) {
		return axios.post(`/users/${userId}/chats`, {
			username: targetUsername 
		});
	},

	// Invia un messaggio in una chat. Supporta testo, foto, e funzionalità avanzate come
    // la risposta a un messaggio specifico (replyTo) e l'inoltro (isForward).
	async sendMessage(userId, chatId, text, photoUrl, replyTo = 0, isForward = false) {
		return axios.post(`/users/${userId}/chats/${chatId}/messages`, {
			text: text,
			photoUrl: photoUrl, 
            replyTo: replyTo,
            isForward: isForward
		});
	},

    // Richiede la cancellazione di un messaggio specifico.
	async deleteMessage(userId, chatId, messageId) {
		return axios.delete(`/users/${userId}/chats/${chatId}/messages/${messageId}`);
	},

    // Legacy: Funzione specifica per l'inoltro, mantenuta per retrocompatibilità ma ora integrata in sendMessage.
    async fowardMessage(userId, chatId, messageId, toUsers) {
        return axios.post(`/users/${userId}/chats/${chatId}/messages/${messageId}/forward`, {
            sentAt: toUsers
        });
    },

    // Aggiunge una reazione (emoticon) a un messaggio specifico.
    async commentMessage(userId, chatId, messageId, emoji) {
        return axios.post(`/users/${userId}/chats/${chatId}/messages/${messageId}/reactions`, {
            emoticon: emoji
        });
    },

    // Rimuove una reazione precedentemente aggiunta a un messaggio.
    async uncommentMessage(userId, chatId, messageId, emoji) {
        return axios.delete(`/users/${userId}/chats/${chatId}/messages/${messageId}/reactions`);
    },

    // Crea un nuovo gruppo specificando nome, lista iniziale di ID membri e foto del gruppo.
    async createGroup(userId, groupName, memberIds, photoUrl) {
        return axios.post(`/groups`, {
            groupname: groupName,
            membersList: memberIds,
            groupPhoto: photoUrl 
        });
    },

    // Aggiunge un nuovo membro a un gruppo esistente utilizzando il suo username.
    async addGroupMember(chatId, username) {
        return axios.post(`/groups/${chatId}/members`, {
            username: username 
        });
    },

    // Aggiunge membri a un gruppo utilizzando una lista di ID numerici.
    async addToGroup(userId, groupId, memberIds) {
        return axios.post(`/groups/${groupId}/members`, {
            membersList: memberIds
        });
    },

    // Permette all'utente di abbandonare un gruppo.
    async leaveGroup(userId, groupId) {
        return axios.delete(`/groups/${groupId}/leave`);
    },

    // Aggiorna il nome visualizzato di un gruppo.
    async setGroupName(userId, groupId, newName) {
        return axios.put(`/groups/${groupId}/name`, {
            groupname: newName
        });
    },

    // Aggiorna l'immagine di un gruppo.
    async setGroupPhoto(userId, groupId, photoUrl) {
        return axios.put(`/groups/${groupId}/photo`, {
            groupPhoto: photoUrl
        });
    },

    async getGroupMembers(groupId) {
        return axios.get(`/groups/${groupId}/members`);
    }

};