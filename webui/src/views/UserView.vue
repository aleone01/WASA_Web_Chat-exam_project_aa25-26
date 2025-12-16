<script>
import api from '@/services/api'

// Componente per la gestione del profilo utente.
// Consente all'utente autenticato di modificare il proprio username
// e di aggiornare l'URL della propria foto profilo.
export default {
    data() {
        return {
            userId: parseInt(sessionStorage.getItem('userId')),
            username: "",
            photoUrl: "",
            msg: null,     
            error: null,   
            loading: false
        }
    },
    methods: {
        // Valida e invia la richiesta di aggiornamento del nome utente.
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
                this.error = "Errore aggiornamento username: " + (e.response?.data?.message || e.toString());
            }
            this.loading = false;
        },

        // Valida e invia la richiesta di aggiornamento della foto profilo.
        async updatePhoto() {
            if (!this.photoUrl) {
                this.error = "Inserisci un URL valido.";
                return;
            }

            this.loading = true;
            this.error = null;
            this.msg = null;

            try {
                await api.setMyPhoto(this.userId, this.photoUrl);
                this.msg = "Foto aggiornata con successo!";
                this.photoUrl = ""; 
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
                        <button @click="updateUsername" class="btn btn-teal" :disabled="loading">
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
                            <label class="form-label text-light">URL Nuova Foto</label>
                            <input v-model="photoUrl" type="url" class="form-control dark-input" placeholder="https://example.com/foto.jpg">
                        </div>
                        <div v-if="photoUrl" class="mb-3 text-center">
                            <img :src="photoUrl" class="rounded-circle border-teal" style="width: 80px; height: 80px; object-fit: cover;" alt="Anteprima">
                        </div>
                        <button @click="updatePhoto" class="btn btn-orange" :disabled="loading">
                            {{ loading ? 'Attendi...' : 'Aggiorna Foto' }}
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.dark-card {
    background-color: #1e1e1e;
    border: 1px solid #333;
}

.dark-input {
    background-color: #2d2d2d;
    border: 1px solid #444;
    color: white;
}
.dark-input:focus {
    background-color: #333;
    color: white;
    border-color: #249EA0;
    box-shadow: 0 0 0 0.25rem rgba(36, 158, 160, 0.25);
}

.btn-teal {
    background-color: #249EA0;
    color: white;
    border: none;
}
.btn-teal:hover {
    background-color: #008083;
    color: white;
}

.btn-orange {
    background-color: #FD5901; /* Arancio/Rosso palette */
    color: white;
    border: none;
}
.btn-orange:hover {
    background-color: #F78104;
    color: white;
}

.border-teal {
    border: 2px solid #249EA0;
}
</style>