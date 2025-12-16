<script setup>
import { RouterLink, RouterView } from 'vue-router'

// App.vue è il componente radice (Root Component) dell'applicazione.
// Definisce il layout strutturale principale, che comprende una barra di navigazione superiore (Header)
// e una barra laterale (Sidebar) per l'accesso rapido alle sezioni Home, Profilo e Logout.
// Il contenuto dinamico delle pagine viene renderizzato all'interno del tag <RouterView>.
</script>

<template>
    <header class="navbar navbar-dark sticky-top flex-md-nowrap p-0 shadow" style="background-color: #005F60;">
        <a class="navbar-brand col-md-3 col-lg-2 me-0 px-3 fs-6" href="#/">WasaText</a>
    </header>

    <div class="container-fluid">
        <div class="row">
            <nav id="sidebarMenu" class="col-md-3 col-lg-2 d-md-block sidebar collapse custom-sidebar-bg">
                <div class="position-sticky pt-3 sidebar-sticky">
                    <h6 class="sidebar-heading d-flex justify-content-between align-items-center px-3 mt-4 mb-1 text-uppercase" style="color: #FAAB36;">
                        <span>Menu</span>
                    </h6>
                    <ul class="nav flex-column">
                        <li class="nav-item">
                            <RouterLink to="/" class="nav-link">
                                <i class="bi bi-house-door"></i> Home: Chats
                            </RouterLink>
                        </li>
                        <li class="nav-item">
                            <RouterLink to="/profile" class="nav-link">
                                <i class="bi bi-person-circle"></i> Modifica Profilo
                            </RouterLink>
                        </li>
                        <li class="nav-item">
                            <RouterLink to="/login" class="nav-link logout-link">
                                <i class="bi bi-box-arrow-right"></i> Logout
                            </RouterLink>
                        </li>
                    </ul>
                </div>
            </nav>

            <main class="col-md-9 ms-sm-auto col-lg-10 px-md-4 main-content">
                <RouterView />
            </main>
        </div>
    </div>
</template>

<style>
/* Global Dark Mode Base */
body {
    background-color: #121212;
    color: #f0f0f0;
}

/* Sidebar styles */
.sidebar {
  position: fixed;
  top: 0;
  bottom: 0;
  left: 0;
  z-index: 100;
  padding: 48px 0 0;
  box-shadow: inset -1px 0 0 rgba(255, 255, 255, .1);
}

/* Custom Dark Sidebar Background */
.custom-sidebar-bg {
    background-color: #1e1e1e;
    border-right: 1px solid #333;
}

.sidebar-sticky {
  position: relative;
  top: 0;
  height: calc(100vh - 48px);
  padding-top: .5rem;
  overflow-x: hidden;
  overflow-y: auto;
}

/* Link Styling */
.sidebar .nav-link {
  font-weight: 500;
  color: #e0e0e0; /* Grigio chiaro per testo normale */
  transition: all 0.2s;
}

.sidebar .nav-link:hover {
    color: #FAAB36; /* Giallo-Arancio all'hover */
    background-color: rgba(255, 255, 255, 0.05);
}

.sidebar .nav-link.active {
  color: #249EA0; /* Teal chiaro per attivo */
  font-weight: bold;
}

.logout-link:hover {
    color: #FD5901 !important; /* Arancio acceso per logout */
}

/* Scrollbar scura */
::-webkit-scrollbar {
    width: 8px;
}
::-webkit-scrollbar-track {
    background: #1e1e1e; 
}
::-webkit-scrollbar-thumb {
    background: #005F60; 
    border-radius: 4px;
}
::-webkit-scrollbar-thumb:hover {
    background: #249EA0; 
}
</style>