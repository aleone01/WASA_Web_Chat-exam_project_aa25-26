<script>
import api from '@/services/api'

// HomeView è il componente principale dell'interfaccia utente.
// Gestisce la visualizzazione a due colonne (lista chat a sinistra, conversazione a destra),
// il polling per i nuovi messaggi, l'invio di messaggi/foto, la creazione di gruppi e la gestione
// delle interazioni come risposte, inoltri e cancellazione messaggi.
export default {
    data: function() {
        return {
            errormsg: null,
            loading: false,
            myId: parseInt(sessionStorage.getItem('userId')) || 0,
            chats: [],              
            messages: [],           
            currentChatId: null,
            currentChatName: "",
            currentChatPhoto: "",
            newMessageText: "",
            newPhotoUrl: "",     
            replyingToMsg: null, 
            refreshInterval: null,
            availableReactions: ['👍', '❤️', '😂', '😮', '😢', '😡'],
            activeReactionMenuId: null,
            showGroupInfoModal: false,
            groupMembers: [],
            newGroupName: "",
            newGroupPhoto: "",
        }
    },
    computed: {
        // Determina se la chat attualmente selezionata è un gruppo basandosi sui metadati della chat.
        isCurrentChatGroup() {
            if (!this.currentChatId) return false;
            const chat = this.chats.find(c => c.id === this.currentChatId);
            return chat ? !!chat.isGroup : false;
        },
        // Recupera lo username dell'utente corrente dalla sessione per identificare le proprie reazioni
        myUsername() { return "Me"; }
    },
    // Lifecycle Hook: all'avvio verifica l'autenticazione, carica i dati iniziali e imposta il polling.
    mounted() {
        if (!this.myId) { this.$router.push('/login'); return; }
        this.refresh();
        this.refreshInterval = setInterval(() => {
            // Aggiorniamo anche la lista chat (refresh) per vedere se ci hanno aggiunto a nuovi gruppi
            this.refresh();
            if (this.currentChatId) {
                api.getConversation(this.myId, this.currentChatId).then(res => {
                    this.messages = res.data.messages || res.data || [];
                });
            }
        }, 3000);
    },
    // Lifecycle Hook: pulisce l'intervallo di polling alla distruzione del componente.
    unmounted() { clearInterval(this.refreshInterval); },
    methods: {
        // Scarica la lista aggiornata delle conversazioni dal server e aggiorna lo stato locale.
        async refresh() {
            try {
                const response = await api.getMyConversations(this.myId);
                this.chats = response.data.chats || response.data || [];
            } catch (e) { console.error(e); }
        },

        // Gestisce la selezione di una chat dalla lista: imposta l'ID corrente, carica i metadati
        // e scarica la cronologia dei messaggi, scorrendo infine verso il basso.
        async selectChat(chatId) {
            this.currentChatId = chatId;
            this.loading = true;
            this.messages = [];
            this.replyingToMsg = null;
            this.activeReactionMenuId = null; // Chiude eventuali menu aperti
            this.showGroupInfoModal = false;

            const selectedChat = this.chats.find(c => c.id === chatId);
            if (selectedChat) {
                this.currentChatName = selectedChat.name;
                this.currentChatPhoto = selectedChat.photoChat;
            }

            try {
                const response = await api.getConversation(this.myId, chatId);
                this.messages = response.data.messages || response.data || [];
                this.$nextTick(() => this.scrollToBottom());
            } catch (e) {
                this.errormsg = "Errore: " + e.toString();
            }
            this.loading = false;
        },


        // Apre il modale e carica i dati del gruppo
        async openGroupInfo() {
            const chat = this.chats.find(c => c.id === this.currentChatId);
            if (!chat || !chat.isGroup) return;

            this.showGroupInfoModal = true;
            this.newGroupName = this.currentChatName;
            this.newGroupPhoto = this.currentChatPhoto;
            await this.loadGroupMembers();
        },

        // Carica la lista membri dal backend
        async loadGroupMembers() {
            try {
                const res = await api.getGroupMembers(this.currentChatId);
                this.groupMembers = res.data.membersList || [];
            } catch(e) {
                console.error("Errore caricamento membri:", e);
            }
        },

        // Aggiunge un membro (dal modale)
        async addMemberToGroup() {
            const username = prompt("Inserisci lo username da aggiungere:");
            if (!username) return;
            try {
                await api.addGroupMember(this.currentChatId, username);
                await this.loadGroupMembers(); // Ricarica la lista visiva
                alert("Membro aggiunto!");
            } catch (e) {
                alert("Errore: " + (e.response?.data?.message || e.toString()));
            }
        },

        // Abbandona il gruppo
        async leaveGroup() {
            if (!confirm("Sei sicuro di voler abbandonare il gruppo?")) return;
            try {
                await api.leaveGroup(this.myId, this.currentChatId);
                this.showGroupInfoModal = false;
                this.currentChatId = null;
                await this.refresh(); // Aggiorna la lista chat (il gruppo sparirà)
            } catch (e) {
                alert("Errore uscita: " + e.toString());
            }
        },

        // Salva modifiche nome/foto
        async saveGroupSettings() {
            try {
                if (this.newGroupName !== this.currentChatName) {
                    await api.setGroupName(this.myId, this.currentChatId, this.newGroupName);
                    this.currentChatName = this.newGroupName;
                }
                if (this.newGroupPhoto !== this.currentChatPhoto) {
                    await api.setGroupPhoto(this.myId, this.currentChatId, this.newGroupPhoto);
                    this.currentChatPhoto = this.newGroupPhoto;
                }
                alert("Gruppo aggiornato!");
                this.refresh(); // Aggiorna la sidebar
            } catch (e) {
                alert("Errore aggiornamento: " + e.toString());
            }
        },

        // Imposta un messaggio specifico come target per una risposta (Reply).
        setReply(msg) {
            this.replyingToMsg = msg;
            this.$nextTick(() => this.$refs.inputField.focus());
        },

        // Annulla la modalità di risposta.
        cancelReply() {
            this.replyingToMsg = null;
        },

        // Cerca e restituisce un frammento del messaggio originale a cui si sta rispondendo
        // per visualizzarlo nell'interfaccia utente.
        getReplySnippet(replyId) {
            if (!replyId) return { found: false, text: 'Messaggio non disponibile', authorId: null };
            
            const parent = this.messages.find(m => m.id === replyId);
            if (parent) {
                const snippet = parent.text ? parent.text : (parent.photoUrl ? '📷 Foto' : '...');
                return { found: true, text: snippet, authorId: parent.sentBy };
            }
            return { found: false, text: 'Messaggio non disponibile', authorId: null };
        },

        // Invia un nuovo messaggio (testo e/o foto) alla chat corrente, gestendo anche eventuali risposte.
        async sendMessage() {
            if (!this.newMessageText && !this.newPhotoUrl) return;
            try {
                const replyToId = this.replyingToMsg ? this.replyingToMsg.id : 0;
                await api.sendMessage(this.myId, this.currentChatId, this.newMessageText, this.newPhotoUrl, replyToId, false);
                this.newMessageText = "";
                this.newPhotoUrl = "";
                this.replyingToMsg = null;
                await this.selectChat(this.currentChatId);
                this.refresh(); 
            } catch (e) {
                this.errormsg = "Errore invio: " + e.toString();
            }
        },

        // Inoltra un messaggio esistente a un altro utente creando (se necessario) una nuova conversazione.
        async forwardMsg(msg) {
            const targetUsername = prompt(`A quale utente (username) vuoi inoltrare questo messaggio?`);
            if (!targetUsername) return;
            try {
                const res = await api.createConversation(this.myId, targetUsername);
                const targetChatId = res.data.id || res.data.Id;
                await api.sendMessage(this.myId, targetChatId, msg.text, msg.photoUrl, 0, true);
                alert("Messaggio inoltrato!");
                this.refresh();
            } catch (e) {
                alert("Errore inoltro: " + e.toString());
            }
        },

        // Scorre automaticamente il contenitore dei messaggi fino all'ultimo elemento.
        scrollToBottom() {
            const container = this.$refs.messageContainer;
            if (container) container.scrollTop = container.scrollHeight;
        },

        // Richiede la cancellazione di un messaggio previo conferma dell'utente.
        async deleteMsg(messageId) {
            if(!confirm("Eliminare messaggio?")) return;
            try {
                await api.deleteMessage(this.myId, this.currentChatId, messageId);
                await this.selectChat(this.currentChatId); 
            } catch (e) { alert(e.toString()); }
        },

        // Gestisce il flusso di creazione di un nuovo gruppo: richiede nome e membri, crea il gruppo
        // e aggiunge i partecipanti iterativamente.
        async createGroup() {
            const name = prompt("Nome gruppo:");
            if (!name) return;

            const membersStr = prompt("Inserisci gli username dei membri separati da virgola (es: luca, marco):");
            
            try { 
                // Crea il gruppo (solo con il creatore inizialmente)
                const res = await api.createGroup(this.myId, name, [this.myId], "");
                const newGroupId = res.data.id || res.data.Id;

                // Se ci sono altri membri, aggiungili uno alla volta
                if (membersStr) {
                    const members = membersStr.split(',').map(s => s.trim()).filter(s => s);
                    for (const member of members) {
                        try {
                            await api.addGroupMember(newGroupId, member);
                        } catch (e) {
                            console.error("Errore aggiunta membro " + member, e);
                        }
                    }
                }

                await this.refresh(); 
                alert("Gruppo creato!");
            } catch (e) { 
                alert("Errore creazione gruppo: " + e.toString()); 
            }
        },

        // Avvia una nuova chat privata chiedendo lo username del destinatario.
        async startConversation() {
            const otherUsername = prompt("Username:");
            if (!otherUsername) return;
            try {
                const response = await api.createConversation(this.myId, otherUsername);
                await this.refresh();
                const chatId = response.data.id || response.data.Id; 
                if (chatId) await this.selectChat(chatId);
            } catch (e) { alert(e.toString()); }
        },

        // Formatta una data ISO in una stringa leggibile (Giorno/Mese Ore:Minuti).
        formatDate(isoString) {
            if (!isoString || isoString.startsWith('0001')) return '';
            const d = new Date(isoString);
            return d.toLocaleString([], { 
                day: '2-digit', 
                month: '2-digit', 
                hour: '2-digit', 
                minute: '2-digit' 
            });
        },

        // Gestisce l'aggiunta di una reazione chiamando l'API.
        async reactToMsg(msgId, emoji) {
            try {
                await api.commentMessage(this.myId, this.currentChatId, msgId, emoji);
                this.activeReactionMenuId = null; // Chiude il menu dopo la scelta
                // Aggiorna la conversazione per vedere la nuova reazione
                const res = await api.getConversation(this.myId, this.currentChatId);
                this.messages = res.data.messages || res.data || [];
            } catch (e) {
                console.error("Errore reazione:", e);
                this.activeReactionMenuId = null;
            }
        },

        // Rimuove la propria reazione da un messaggio.
        async removeReaction(msgId) {
            try {
                await api.uncommentMessage(this.myId, this.currentChatId, msgId);
                const res = await api.getConversation(this.myId, this.currentChatId);
                this.messages = res.data.messages || res.data || [];
            } catch (e) {
                console.error("Errore rimozione reazione:", e);
            }
        },

        // Toggle per mostrare/nascondere il menu delle emoticon per un dato messaggio.
        toggleReactionMenu(msgId) {
            if (this.activeReactionMenuId === msgId) {
                this.activeReactionMenuId = null;
            } else {
                this.activeReactionMenuId = msgId;
            }
        }
    }
}
</script>

<template>
  <div class="container-fluid vh-100 p-0 d-flex flex-column font-sans bg-dark-theme position-relative">
    <div class="d-flex justify-content-between align-items-center py-2 px-3 border-bottom-dark header-bg" style="flex: 0 0 auto;">
      <h4 class="m-0 fw-bold" style="color: #249EA0;">WasaText</h4>
      <div class="btn-group">
        <button class="btn btn-sm btn-teal" @click="startConversation">Nuova Chat</button>
        <button class="btn btn-sm btn-outline-orange" @click="createGroup">Crea Gruppo</button>
      </div>
    </div>

    <div v-if="errormsg" class="alert alert-danger mx-3 mt-2 mb-0 py-2">{{ errormsg }}</div>

    <div class="row g-0 flex-grow-1" style="min-height: 0; overflow: hidden;">
      <div class="col-md-4 col-lg-3 border-end-dark d-flex flex-column h-100 sidebar-bg">
        <div class="overflow-auto flex-grow-1">
          <ul class="list-group list-group-flush">
            <li
              v-for="chat in chats" :key="chat.id" 
              class="list-group-item list-group-item-action cursor-pointer d-flex align-items-center py-3 border-bottom-dark chat-item"
              :class="{ 'active-chat': currentChatId === chat.id }"
              @click="selectChat(chat.id)"
            >
              <div class="me-3">
                <img v-if="chat.photoChat" :src="chat.photoChat" class="rounded-circle border-teal" style="width: 45px; height: 45px; object-fit: cover;">
                <div v-else class="rounded-circle bg-dark-circle d-flex align-items-center justify-content-center fw-bold" style="width: 45px; height: 45px; color: #FAAB36;">
                  {{ (chat.name || '?').charAt(0).toUpperCase() }}
                </div>
              </div>
              <div class="flex-grow-1 overflow-hidden">
                <div class="d-flex w-100 justify-content-between">
                  <h6 class="mb-0 text-truncate fw-semibold text-light">{{ chat.name || 'Chat ' + chat.id }}</h6>
                  <small class="text-muted" style="font-size: 0.8rem;">{{ formatDate(chat.lastMessage) }}</small>
                </div>
                <p class="mb-0 small text-truncate text-muted mt-1">{{ chat.snippetText || '...' }}</p>
              </div>
            </li>
          </ul>
        </div>
      </div>

      <div class="col-md-8 col-lg-9 d-flex flex-column h-100 bg-chat-pattern position-relative">
        <div v-if="!currentChatId" class="d-flex align-items-center justify-content-center h-100 text-muted">
          <h4>Seleziona una chat</h4>
        </div>
                
        <div v-else class="d-flex flex-column h-100 w-100">
          <div class="d-flex align-items-center px-3 py-2 header-bg border-bottom-dark shadow-sm" style="flex: 0 0 auto;">
            <img v-if="currentChatPhoto" :src="currentChatPhoto" class="rounded-circle border-teal me-2" style="width: 35px; height: 35px; object-fit: cover;">
            <div v-else class="rounded-circle bg-dark-circle d-flex align-items-center justify-content-center fw-bold me-2" style="width: 35px; height: 35px; color: #FAAB36;">
              {{ (currentChatName || '?').charAt(0).toUpperCase() }}
            </div>
            <h5 class="m-0 text-light flex-grow-1">{{ currentChatName }}</h5>

            <button v-if="isCurrentChatGroup" class="btn btn-sm btn-outline-orange ms-2" title="Info Gruppo" @click="openGroupInfo">
              ℹ️ Info
            </button>
          </div>

          <div ref="messageContainer" class="flex-grow-1 overflow-auto p-3">
            <div v-for="msg in messages" :key="msg.id" class="d-flex mb-2 w-100" :class="{ 'justify-content-end': msg.sentBy == myId }">
              <div
                class="card border-0 shadow-sm msg-bubble position-relative" 
                :class="{ 'msg-sent': msg.sentBy == myId, 'msg-received': msg.sentBy != myId }"
              >
                <div class="card-body p-2">
                  <div v-if="isCurrentChatGroup && msg.sentBy != myId" class="fw-bold mb-1" style="font-size: 0.75rem; color: #FAAB36;">
                    {{ msg.senderName || 'Utente ' + msg.sentBy }}
                  </div>

                  <div v-if="msg.isForward" class="mb-1 fst-italic" style="font-size: 0.75rem; color: #e0e0e0;">
                    <span class="me-1">↪</span>Inoltrato
                  </div>
                  <div v-if="msg.replyTo" class="mb-2 p-2 rounded border-start border-4 reply-box" :class="msg.sentBy == myId ? 'reply-sent' : 'reply-received'">
                    <div class="d-flex flex-column">
                      <template v-if="getReplySnippet(msg.replyTo).found">
                        <small class="fw-bold" style="font-size: 0.7rem; color: #FAAB36;">
                          {{ getReplySnippet(msg.replyTo).authorId == myId ? 'Tu' : 'Utente' }}
                        </small>
                        <span class="text-truncate small opacity-75">{{ getReplySnippet(msg.replyTo).text }}</span>
                      </template>
                      <template v-else><small class="text-muted fst-italic">Msg originale non disponibile</small></template>
                    </div>
                  </div>
                  <div v-if="msg.photoUrl" class="mb-2"><img :src="msg.photoUrl" class="img-fluid rounded" style="max-height: 300px;"></div>
                  <p class="mb-1 text-break" style="white-space: pre-wrap;">{{ msg.text }}</p>
                                    
                  <div v-if="msg.reactions && msg.reactions.length > 0" class="d-flex flex-wrap gap-1 mb-1 mt-1 ms-1">
                    <span
                      v-for="(reaction, idx) in msg.reactions" :key="idx" 
                      class="badge bg-dark-circle text-light border border-secondary rounded-pill d-flex align-items-center px-2 py-1"
                      style="font-size: 0.85rem; cursor: pointer; user-select: none;"
                      :title="'Reazione di ' + reaction.username" @click="removeReaction(msg.id)"
                    >
                      {{ reaction.emoticon }} <span style="font-size: 0.6rem; margin-left: 4px; color: #aaa;">{{ reaction.username }}</span>
                    </span>
                  </div>

                  <div class="d-flex justify-content-end align-items-center mt-1 text-nowrap" style="font-size: 0.7rem; opacity: 0.7;">
                    <span class="me-2">{{ formatDate(msg.sentAt) }}</span>
                    <span v-if="msg.sentBy == myId" class="me-2">
                      <span v-if="msg.checkmark" style="color: #FAAB36; font-weight: bold;">✓✓</span>
                      <span v-else>✓</span>
                    </span>
                    <div class="d-flex gap-2 bg-dark rounded px-1 action-buttons position-relative">
                      <span class="cursor-pointer" title="Reagisci" @click.stop="toggleReactionMenu(msg.id)">😀</span>
                      <span class="cursor-pointer" title="Rispondi" @click.stop="setReply(msg)">↩️</span>
                      <span class="cursor-pointer" title="Inoltra" @click.stop="forwardMsg(msg)">➡️</span>
                      <span v-if="msg.sentBy == myId" class="cursor-pointer text-danger" title="Elimina" @click.stop="deleteMsg(msg.id)">🗑️</span>
                      <div
                        v-if="activeReactionMenuId === msg.id" class="position-absolute bg-dark shadow rounded p-2 d-flex gap-2 border border-secondary" 
                        :style="{ bottom: '100%', zIndex: 1000, right: msg.sentBy == myId ? '0' : 'auto', left: msg.sentBy != myId ? '0' : 'auto' }"
                      >
                        <button v-for="emoji in availableReactions" :key="emoji" class="btn btn-sm btn-dark p-1 border-0" style="font-size: 1.2rem;" @click.stop="reactToMsg(msg.id, emoji)">{{ emoji }}</button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="input-area border-top-dark p-2" style="flex: 0 0 auto;">
            <div v-if="replyingToMsg" class="mb-2 p-2 bg-dark rounded border-start border-4 border-orange d-flex justify-content-between align-items-center">
              <div class="overflow-hidden">
                <small class="text-orange fw-bold d-block">Rispondendo a:</small>
                <small class="text-truncate d-block text-muted">{{ replyingToMsg.text || '📷 Foto' }}</small>
              </div>
              <button class="btn btn-close btn-close-white btn-sm ms-2" @click="cancelReply" />
            </div>
            <div class="input-group">
              <input ref="inputField" v-model="newMessageText" type="text" class="form-control dark-input" placeholder="Scrivi un messaggio..." @keyup.enter="sendMessage">
              <button class="btn btn-teal" @click="sendMessage">➤</button>
            </div>
            <div class="mt-2">
              <input v-model="newPhotoUrl" type="url" class="form-control form-control-sm border-0 dark-input-sm" placeholder="URL Immagine (opzionale)">
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showGroupInfoModal" class="modal-overlay d-flex align-items-center justify-content-center">
      <div class="card shadow p-0 modal-content dark-modal" style="width: 400px; max-width: 90%;">
        <div class="card-header bg-dark-teal text-white d-flex justify-content-between align-items-center">
          <h5 class="mb-0">Gestione Gruppo</h5>
          <button class="btn-close btn-close-white" @click="showGroupInfoModal = false" />
        </div>
        <div class="card-body overflow-auto" style="max-height: 70vh;">
          <div class="mb-3">
            <label class="form-label fw-bold text-light">Nome Gruppo</label>
            <input v-model="newGroupName" class="form-control dark-input mb-2" placeholder="Nome gruppo">
            <label class="form-label fw-bold text-light">Foto Gruppo (URL)</label>
            <input v-model="newGroupPhoto" class="form-control dark-input mb-2" placeholder="URL Foto">
            <button class="btn btn-sm btn-outline-teal w-100" @click="saveGroupSettings">Salva Modifiche</button>
          </div>
          <hr class="border-secondary">

          <h6 class="fw-bold mb-2 text-orange">Membri ({{ groupMembers.length }})</h6>
          <ul class="list-group list-group-flush mb-3">
            <li v-for="m in groupMembers" :key="m.id" class="list-group-item d-flex align-items-center px-0 bg-transparent text-light border-secondary">
              <div
                class="rounded-circle bg-dark-circle d-flex align-items-center justify-content-center me-2 fw-bold" 
                style="width: 32px; height: 32px; font-size: 0.8rem; color: #FAAB36;"
              >
                {{ m.profilePhoto ? '' : m.username.charAt(0).toUpperCase() }}
                <img v-if="m.profilePhoto" :src="m.profilePhoto" class="rounded-circle w-100 h-100 object-fit-cover">
              </div>
              <span>{{ m.username }}</span>
              <span v-if="m.id === myId" class="badge bg-teal ms-auto">Tu</span>
            </li>
          </ul>
                    
          <button class="btn btn-sm btn-success w-100 mb-2" @click="addMemberToGroup">Aggiungi Membro</button>
          <hr class="border-secondary">
          <button class="btn btn-sm btn-danger w-100" @click="leaveGroup">Abbandona Gruppo</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* GENERAL DARK THEME UTILS */
.font-sans { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; }
.vh-100 { height: 100vh !important; }
.cursor-pointer { cursor: pointer; }

/* Colors */
.bg-dark-theme { background-color: #121212; color: #f0f0f0; }
.header-bg { background-color: #005F60; border-bottom: 1px solid #004445; }
.sidebar-bg { background-color: #1e1e1e; }
.bg-chat-pattern { background-color: #1a1a1a; }
.border-bottom-dark { border-bottom: 1px solid #333; }
.border-end-dark { border-right: 1px solid #333; }
.border-teal { border: 2px solid #249EA0; }
.text-orange { color: #FD5901; }
.text-light { color: #f0f0f0; }

/* Buttons */
.btn-teal { background-color: #249EA0; color: white; border: none; }
.btn-teal:hover { background-color: #008083; }
.btn-outline-teal { border: 1px solid #249EA0; color: #249EA0; }
.btn-outline-teal:hover { background-color: #249EA0; color: white; }
.btn-outline-orange { border: 1px solid #FD5901; color: #FD5901; }
.btn-outline-orange:hover { background-color: #FD5901; color: white; }
.bg-teal { background-color: #008083; }
.bg-dark-teal { background-color: #005F60; }

/* Sidebar Items */
.chat-item { background-color: #1e1e1e; color: #ddd; border-bottom: 1px solid #b1a1a1; }
.chat-item:hover { background-color: #d6c9c9; }
.active-chat { background-color: #2d2d2d; border-left: 4px solid #FD5901; }
.bg-dark-circle { background-color: #333; }

/* Messages */
.msg-bubble { position: relative; width: fit-content; max-width: 75%; min-width: 100px; }
.msg-sent { background-color: #008083; color: white; border-radius: 8px 0 8px 8px; }
.msg-received { background-color: #2d2d2d; color: white; border-radius: 0 8px 8px 8px; }

/* Reply Box */
.reply-box { background-color: rgba(0,0,0,0.2); font-size: 0.85rem; }
.reply-sent { border-color: #FAAB36 !important; } /* Yellow-Orange */
.reply-received { border-color: #249EA0 !important; }

/* Inputs */
.input-area { background-color: #1e1e1e; border-top: 1px solid #333; }
.dark-input { background-color: #2d2d2d; border: 1px solid #444; color: white; }
.dark-input:focus { background-color: #333; color: white; border-color: #249EA0; outline: none; }
.dark-input-sm { background-color: #2d2d2d; color: #aaa; }

/* Modal */
.modal-overlay { position: absolute; top: 0; left: 0; right: 0; bottom: 0; background-color: rgba(0,0,0,0.7); z-index: 2000; }
.dark-modal { background-color: #1e1e1e; border: 1px solid #333; color: #f0f0f0; }
.modal-content { animation: fadeIn 0.2s; }

@keyframes fadeIn { from { opacity: 0; transform: scale(0.95); } to { opacity: 1; transform: scale(1); } }

/* Scrollbars */
::-webkit-scrollbar { width: 6px; }
::-webkit-scrollbar-thumb { background: #005F60; border-radius: 3px; }
::-webkit-scrollbar-thumb:hover { background: #249EA0; }
::-webkit-scrollbar-track { background: #1e1e1e; }
</style>