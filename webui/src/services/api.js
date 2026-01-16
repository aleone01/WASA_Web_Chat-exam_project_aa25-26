import axios from "./axios";

// Modulo API Service.
// Questo oggetto esporta tutte le funzioni necessarie per interagire con il backend.
// Astrae le chiamate HTTP dirette, offrendo ai componenti Vue metodi semantici (es. doLogin, sendMessage)
// e gestendo la costruzione dei payload JSON o Multipart richiesti dalle API REST in Go.
export default {
	
    // Effettua il login o la registrazione implicita.
	async doLogin(username) {
		return axios.post("/login", { username });
	},

    // Aggiorna lo username dell'utente.
	async setMyUserName(userId, newUsername) {
		return axios.put(`/users/${userId}/username`, { username: newUsername });
	},
	
    // Aggiorna la foto profilo.
    // Utilizza FormData perché stiamo inviando un file binario, non un JSON.
	async setMyPhoto(userId, photoFile) {
        let formData = new FormData();
        formData.append("photo", photoFile);
		return axios.put(`/users/${userId}/photo`, formData, {
            headers: { 'Content-Type': 'multipart/form-data' }
        });
	},

    // Recupera la lista delle chat (home screen).
	async getMyConversations(userId) {
		return axios.get(`/users/${userId}/chats`);
	},

    // Recupera i dettagli e i messaggi di una singola chat.
	async getConversation(userId, chatId) {
		return axios.get(`/users/${userId}/chats/${chatId}`);
	},

    // Crea una nuova conversazione privata (o restituisce quella esistente).
	async createConversation(userId, targetUsername) {
		return axios.post(`/users/${userId}/chats`, {
			username: targetUsername 
		});
	},

    // Invia un messaggio in una chat.
    // Gestisce un payload misto (testo + file opzionale) tramite FormData.
    // Supporta anche i metadati per le risposte (replyTo) e gli inoltri (isForward).
	async sendMessage(userId, chatId, text, photoFile, replyTo = 0, isForward = false) {
        let formData = new FormData();
        formData.append("text", text || "");
        if (photoFile) {
            formData.append("file", photoFile);
        }
        formData.append("replyTo", replyTo);
        formData.append("isForward", isForward);

		return axios.post(`/users/${userId}/chats/${chatId}/messages`, formData, {
            headers: { 'Content-Type': 'multipart/form-data' }
        });
	},

    // Inoltra un messaggio esistente a una lista di target (broadcast forwarding).
    async forwardMessage(userId, chatId, messageId, toUsers) {
        return axios.post(`/users/${userId}/chats/${chatId}/messages/${messageId}/forward`, {
            targets: toUsers
        });
    },

    // Cancella un messaggio specifico.
	async deleteMessage(userId, chatId, messageId) {
		return axios.delete(`/users/${userId}/chats/${chatId}/messages/${messageId}`);
	},

    // Aggiunge una reazione (emoji) a un messaggio.
    async commentMessage(userId, chatId, messageId, emoji) {
        return axios.post(`/users/${userId}/chats/${chatId}/messages/${messageId}/reactions`, {
            emoticon: emoji
        });
    },

    // Rimuove una reazione da un messaggio.
    async uncommentMessage(userId, chatId, messageId, emoji) {
        return axios.delete(`/users/${userId}/chats/${chatId}/messages/${messageId}/reactions`);
    },

    // Crea un nuovo gruppo.
    // Richiede multipart/form-data per supportare l'upload opzionale della foto gruppo.
    // La lista dei membri viene serializzata in JSON stringa per essere inviata come campo form.
    async createGroup(userId, groupName, memberIds, photoFile) {
        let formData = new FormData();
        formData.append("groupname", groupName);
        formData.append("membersList", JSON.stringify(memberIds)); // Serializza array in stringa per il form
        if (photoFile) {
            formData.append("groupPhotoFile", photoFile);
        }

        return axios.post(`/groups`, formData, {
            headers: { 'Content-Type': 'multipart/form-data' }
        });
    },

    // Aggiunge un membro a un gruppo esistente.
    async addGroupMember(chatId, username) {
        return axios.post(`/groups/${chatId}/members`, {
            username: username 
        });
    },

    // Abbandona un gruppo.
    async leaveGroup(userId, groupId) {
        return axios.delete(`/groups/${groupId}/leave`);
    },

    // Aggiorna il nome del gruppo.
    async setGroupName(userId, groupId, newName) {
        return axios.put(`/groups/${groupId}/name`, {
            groupname: newName
        });
    },

    // Aggiorna la foto del gruppo.
    async setGroupPhoto(userId, groupId, photoFile) {
        let formData = new FormData();
        formData.append("groupPhotoFile", photoFile);
        return axios.put(`/groups/${groupId}/photo`, formData, {
            headers: { 'Content-Type': 'multipart/form-data' }
        });
    },

    // Ottiene la lista dei membri di un gruppo.
    async getGroupMembers(groupId) {
        return axios.get(`/groups/${groupId}/members`);
    }
};