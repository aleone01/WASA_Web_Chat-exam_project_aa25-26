<script>
import api from '@/services/api'

export default {
    data() {
        return {
            userId: parseInt(sessionStorage.getItem('userId')),
            username: sessionStorage.getItem('username') || "", 
            currentPhoto: sessionStorage.getItem('userPhoto'),  
            photoFile: null,
            msg: null,     
            error: null,   
            loading: false
        }
    },

    methods: {

        // Helper per visualizzare l'immagine base64
        getImageSrc(base64Data) {
            if (!base64Data) return null;
            if (base64Data.startsWith('data:')) return base64Data;
            return `data:image/jpeg;base64,${base64Data}`;
        },

        // Helper per notificare tutte le chat (messaggio di sistema)
        async notifyAllChats(messageContent) {
            try {
                const res = await api.getMyConversations(this.userId);
                const chats = res.data.chats || res.data || [];
                const systemText = `[INFO]: ${messageContent}`;
                
                const promises = chats.map(chat => 
                    api.sendMessage(this.userId, chat.id, systemText, null)
                );
                await Promise.all(promises);
            } catch (e) {
                console.error("Errore invio notifica chat", e);
            }
        },

        async updateUsername() {
          
            if (!this.username || this.username.length < 3 || this.username.length > 16) {
                this.error = "L'username deve essere tra 3 e 16 caratteri.";
                return;
            }

           
            const oldUsername = sessionStorage.getItem('username');
            if (this.username === oldUsername) {
                this.msg = "Nessuna modifica effettuata all'username.";
                this.error = null;
                return;
            }

            this.loading = true;
            this.error = null;
            this.msg = null;

            try {
                await api.setMyUserName(this.userId, this.username);
                
                sessionStorage.setItem('username', this.username);
                await this.notifyAllChats(`L'utente ha cambiato username in "${this.username}"`);

                this.msg = "Username aggiornato con successo!";
            } catch (e) {
            
                const errMsg = e.response?.data?.message || e.response?.data || "";
                const errStr = String(errMsg).toLowerCase();
                const status = e.response?.status;

                // Se l'errore è un conflitto (409), un errore server generico (500),
                // o se il messaggio contiene parole chiave di errore/duplicazione.
                if (status === 409 || 
                    status === 500 || 
                    errStr.includes("taken") || 
                    errStr.includes("exist") || 
                    errStr.includes("duplicate") || 
                    errStr.includes("constraint") ||
                    errStr.includes("errore")) { 
                    
                    this.error = `L'username "${this.username}" è già in uso. Scegline un altro.`;
                } else {
                    // Fallback per altri errori (es. rete)
                    this.error = "Si è verificato un problema: " + (errMsg || e.message);
                }
            }
            this.loading = false;
        },

        // Gestione del cambio file
        onFileChange(e) {
            const files = e.target.files || e.dataTransfer.files;
            if (!files.length) return;
            this.photoFile = files[0];
        },

        async updatePhoto() {
            if (!this.photoFile) {
                this.error = "Seleziona un file prima di caricare.";
                return;
            }

            this.loading = true;
            this.error = null;
            this.msg = null;

            try {

                await api.setMyPhoto(this.userId, this.photoFile);
                const reader = new FileReader();
                reader.onload = async (e) => {
                    const base64String = e.target.result;
                    this.currentPhoto = base64String;
                    const rawBase64 = base64String.split(',')[1];
                    
                    sessionStorage.setItem('userPhoto', rawBase64);
                    localStorage.setItem(`wasa_photo_${this.userId}`, rawBase64);

                    const currentName = sessionStorage.getItem('username') || "L'utente";
                    await this.notifyAllChats(`${currentName} ha aggiornato la foto profilo`);
                };
                reader.readAsDataURL(this.photoFile);

                this.msg = "Foto aggiornata con successo!";
                this.photoFile = null; 
                if (this.$refs.fileInput) this.$refs.fileInput.value = "";
                
            } catch (e) {
                this.error = "Errore aggiornamento foto: " + (e.response?.data?.message || e.toString());
            }
            this.loading = false;
        }
    }
}
</script>

<template>
  <div class="container mt-4">
    <h2 class="mb-4" style="color: #249EA0;">Il mio Profilo</h2>

    <div v-if="error" class="alert alert-danger" style="background-color: #3e1a1a; border-color: #FD5901; color: #ffbaba;">{{ error }}</div>
    <div v-if="msg" class="alert alert-success" style="background-color: #1a3e3e; border-color: #249EA0; color: #aaffff;">{{ msg }}</div>

    <div class="row">
      <div class="col-md-6 mb-3">
        <div class="card shadow-sm dark-card">
          <div class="card-header text-white" style="background-color: #005F60;">
            Cambia Username
          </div>
          <div class="card-body">
            <div class="mb-3">
              <label class="form-label text-light">Username Attuale</label>
              <input v-model="username" type="text" class="form-control dark-input" placeholder="Min 3, Max 16 caratteri" minlength="3" maxlength="16">
            </div>
            <button class="btn btn-teal" :disabled="loading" @click="updateUsername">
              {{ loading ? 'Attendi...' : 'Aggiorna Username' }}
            </button>
          </div>
        </div>
      </div>

      <div class="col-md-6 mb-3">
        <div class="card shadow-sm dark-card">
          <div class="card-header text-white" style="background-color: #FD5901;">
            Cambia Foto Profilo
          </div>
          <div class="card-body text-center">
            <div class="mb-3">
              <p class="text-light mb-1">Foto Attuale:</p>
              <img v-if="getImageSrc(currentPhoto)" :src="getImageSrc(currentPhoto)" class="rounded-circle border border-light" style="width: 100px; height: 100px; object-fit: cover;">
              <div v-else class="rounded-circle bg-secondary d-inline-flex align-items-center justify-content-center text-white" style="width: 100px; height: 100px;">
                <span style="font-size: 2rem;">{{ (username || 'U').charAt(0).toUpperCase() }}</span>
              </div>
            </div>

            <div class="mb-3 text-start">
              <label class="form-label text-light">Carica Nuova Foto</label>
              <input ref="fileInput" type="file" class="form-control dark-input" accept="image/*" @change="onFileChange">
            </div>
            <button class="btn btn-orange w-100" :disabled="loading" @click="updatePhoto">
              {{ loading ? 'Attendi...' : 'Carica Foto' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dark-card { background-color: #1e1e1e; border: 1px solid #333; }
.dark-input { background-color: #2d2d2d; border: 1px solid #444; color: white; }
.dark-input:focus { background-color: #333; color: white; border-color: #249EA0; box-shadow: 0 0 0 0.25rem rgba(36, 158, 160, 0.25); }
.btn-teal { background-color: #249EA0; color: white; border: none; }
.btn-teal:hover { background-color: #008083; color: white; }
.btn-orange { background-color: #FD5901; color: white; border: none; }
.btn-orange:hover { background-color: #F78104; color: white; }
</style>