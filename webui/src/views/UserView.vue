<script>
import api from '@/services/api'

export default {
    data() {
        return {
            userId: parseInt(sessionStorage.getItem('userId')),
            username: "",
            photoFile: null,
            msg: null,     
            error: null,   
            loading: false
        }
    },
    methods: {
        async updateUsername() {
            if (!this.username || this.username.length < 3 || this.username.length > 16) {
                this.error = "L'username deve essere tra 3 e 16 caratteri.";
                return;
            }

            this.loading = true;
            this.error = null;
            this.msg = null;

            try {
                await api.setMyUserName(this.userId, this.username);
                this.msg = "Username aggiornato con successo!";
                this.username = ""; 
            } catch (e) {
                if (e.response && e.response.data) {
                    alert("Errore: " + e.response.data); 
                } else {
                    alert("Errore aggiornamento username: Impossibile contattare il server.");
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
                this.error = "Seleziona un file.";
                return;
            }

            this.loading = true;
            this.error = null;
            this.msg = null;

            try {
                await api.setMyPhoto(this.userId, this.photoFile);
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
              <label class="form-label text-light">Nuovo Username</label>
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
          <div class="card-body">
            <div class="mb-3">
              <label class="form-label text-light">Carica Nuova Foto</label>
              <input ref="fileInput" type="file" class="form-control dark-input" accept="image/*" @change="onFileChange">
            </div>
            <button class="btn btn-orange" :disabled="loading" @click="updatePhoto">
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