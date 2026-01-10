<template>
  <div class="d-flex align-items-center justify-content-center" style="height: 100vh; background-color: #121212;">
    <div class="card shadow p-4 login-card" style="width: 350px;">
      <h3 class="text-center mb-3" style="color: #249EA0;">WasaText Login</h3>
      <form @submit.prevent="doLogin">
        <div class="mb-3">
          <label class="form-label text-light">Username</label>
          <input v-model="username" type="text" class="form-control dark-input" placeholder="Tuo username" minlength="3" maxlength="16" required>
        </div>
        <button type="submit" class="btn w-100 btn-orange">Entra</button>
      </form>
      <div v-if="error" class="alert alert-danger mt-3 py-2" role="alert" style="background-color: #3e1a1a; border-color: #FD5901; color: #ffbaba;">
        {{ error }}
      </div>
    </div>
  </div>
</template>

<script>
import api from '@/services/api';

// Componente per la gestione del Login.
// Fornisce un modulo semplice per inserire lo username.
// Invia la richiesta di autenticazione all'API e, in caso di successo,
// salva il token di sessione (ID utente) e reindirizza alla home page.
export default {
  data() {
    return {
      username: '',
      error: null
    };
  },
  methods: {
    // Esegue la procedura di login. Gestisce la risposta dell'API estraendo
    // l'identificativo utente e salvandolo in sessionStorage.
    async doLogin() {
      try {
        const response = await api.doLogin(this.username);
        
        let userId = response.data.id || response.data.identifier || response.data.userId;

        if (!userId && typeof response.data === 'number') {
             userId = response.data;
        }

        if (!userId) {
            this.error = "Errore: ID utente non trovato nella risposta.";
            console.error("ID non trovato! Struttura ricevuta:", response.data);
            return;
        }
        
        sessionStorage.clear();
        const savedPhoto = localStorage.getItem(`wasa_photo_${userId}`);
        if (savedPhoto) {
            sessionStorage.setItem('userPhoto', savedPhoto);
        }

        sessionStorage.setItem('token', userId); 
        sessionStorage.setItem('userId', userId);
        sessionStorage.setItem('username', this.username);
        
        this.$router.push('/');
      } catch (e) {
        this.error = "Login fallito: " + (e.response?.data?.message || e.message);
      }
    }
  }
};
</script>

<style scoped>
.login-card {
    background-color: #1e1e1e;
    border: 1px solid #005F60;
    color: white;
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

.btn-orange {
    background-color: #FD5901;
    color: white;
    border: none;
    font-weight: bold;
}
.btn-orange:hover {
    background-color: #F78104;
    color: white;
}
</style>