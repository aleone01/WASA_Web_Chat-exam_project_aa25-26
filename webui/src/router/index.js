import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import HomeView from '../views/HomeView.vue'
import UserView from '../views/UserView.vue'

// Configura il router principale dell'applicazione utilizzando la history mode di HTML5.
// Definisce la mappatura tra i percorsi URL (es. /login, /, /profile) e i componenti Vue corrispondenti.
// Include la logica per riutilizzare lo stesso componente (HomeView) sia per la dashboard principale
// che per la visualizzazione di chat specifiche.
const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: LoginView 
    },
    {
      path: '/',
      name: 'home',
      component: HomeView 
    },
    {
        path: '/chats/:chatId',
        name: 'chat',
        component: HomeView 
    },
    {
        path: '/profile',
        name: 'profile',
        component: UserView
    }
  ]
})

// Implementa un navigation guard globale (beforeEach).
// Questa funzione viene eseguita prima di ogni cambio di rotta per verificare se la pagina richiesta
// necessita di autenticazione. Se l'utente non possiede un token valido in session storage e tenta
// di accedere a pagine protette, viene reindirizzato forzatamente alla pagina di login.
router.beforeEach((to, from, next) => {
    const publicPages = ['/login'];
    const authRequired = !publicPages.includes(to.path);
    const loggedIn = sessionStorage.getItem('token');
  
    if (authRequired && !loggedIn) {
      next('/login');
    } else {
      next();
    }
  });

export default router