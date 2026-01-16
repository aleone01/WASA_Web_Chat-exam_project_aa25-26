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

// Componente LoginView.
// Gestisce l'ingresso nell'applicazione. Poiché la registrazione è implicita (se l'utente non esiste, viene creato),
// questo form serve sia per il Sign-In che per il Sign-Up.
export default {
  data() {
    return {
      username: '',
      error: null
    };
  },
  methods: {
    // Esegue la procedura di login. 
    // 1. Chiama l'API.
    // 2. Parsa l'ID utente dalla risposta (gestendo vari formati possibili).
    // 3. Salva i dati essenziali in sessionStorage (persistenza solo per la sessione corrente del browser).
    // 4. Reindirizza alla Home.
    async doLogin() {
      try {
        const response = await api.doLogin(this.username);
        
        // Normalizzazione della risposta per estrarre l'ID utente
        let userId = response.data.id || response.data.identifier || response.data.userId;

        // Fallback per risposte che contengono direttamente l'ID numerico
        if (!userId && typeof response.data === 'number') {
             userId = response.data;
        }

        if (!userId) {
            this.error = "Errore: ID utente non trovato nella risposta.";
            console.error("ID non trovato! Struttura ricevuta:", response.data);
            return;
        }
        
        // Pulizia sessione precedente
        sessionStorage.clear();

        // Tentativo di recupero foto dalla cache locale (LocalStorage) per evitare flash grafici
        const savedPhoto = localStorage.getItem(`wasa_photo_${userId}`);
        if (savedPhoto) {
            sessionStorage.setItem('userPhoto', savedPhoto);
        }

        // Salvataggio credenziali di sessione
        sessionStorage.setItem('token', userId); 
        sessionStorage.setItem('userId', userId);
        sessionStorage.setItem('username', this.username);
        
        // Navigazione alla dashboard
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