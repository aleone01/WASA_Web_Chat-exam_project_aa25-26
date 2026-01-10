import axios from "./axios";

// Modulo API Service.
// Questo oggetto esporta tutte le funzioni necessarie per interagire con il backend.
// Astrae le chiamate HTTP dirette, offrendo ai componenti Vue metodi semantici (es. doLogin, sendMessage)
// e gestendo la costruzione dei payload JSON richiesti dalle API REST in Go.
export default {
	
	async doLogin(username) {
		return axios.post("/login", { username });
	},

	async setMyUserName(userId, newUsername) {
		return axios.put(`/users/${userId}/username`, { username: newUsername });
	},
	
	async setMyPhoto(userId, photoFile) {
        let formData = new FormData();
        formData.append("photo", photoFile);
		return axios.put(`/users/${userId}/photo`, formData, {
            headers: { 'Content-Type': 'multipart/form-data' }
        });
	},

	async getMyConversations(userId) {
		return axios.get(`/users/${userId}/chats`);
	},

	async getConversation(userId, chatId) {
		return axios.get(`/users/${userId}/chats/${chatId}`);
	},

	async createConversation(userId, targetUsername) {
		return axios.post(`/users/${userId}/chats`, {
			username: targetUsername 
		});
	},

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

    async forwardMessage(userId, chatId, messageId, toUsers) {
        return axios.post(`/users/${userId}/chats/${chatId}/messages/${messageId}/forward`, {
            targets: toUsers
        });
    },

	async deleteMessage(userId, chatId, messageId) {
		return axios.delete(`/users/${userId}/chats/${chatId}/messages/${messageId}`);
	},

    async commentMessage(userId, chatId, messageId, emoji) {
        return axios.post(`/users/${userId}/chats/${chatId}/messages/${messageId}/reactions`, {
            emoticon: emoji
        });
    },

    async uncommentMessage(userId, chatId, messageId, emoji) {
        return axios.delete(`/users/${userId}/chats/${chatId}/messages/${messageId}/reactions`);
    },

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

    async addGroupMember(chatId, username) {
        return axios.post(`/groups/${chatId}/members`, {
            username: username 
        });
    },

    async addToGroup(groupId, username) {
        return axios.post(`/groups/${groupId}/members`, {
            username: username 
        });
    },

    async leaveGroup(userId, groupId) {
        return axios.delete(`/groups/${groupId}/leave`);
    },

    async setGroupName(userId, groupId, newName) {
        return axios.put(`/groups/${groupId}/name`, {
            groupname: newName
        });
    },

    async setGroupPhoto(userId, groupId, photoFile) {
        let formData = new FormData();
        formData.append("groupPhotoFile", photoFile);
        return axios.put(`/groups/${groupId}/photo`, formData, {
            headers: { 'Content-Type': 'multipart/form-data' }
        });
    },

    async getGroupMembers(groupId) {
        return axios.get(`/groups/${groupId}/members`);
    }
};