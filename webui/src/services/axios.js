import axios from "axios";

// Crea e configura l'istanza globale di Axios per la gestione delle chiamate HTTP.
// Imposta l'URL di base (definito nelle variabili d'ambiente) e un timeout di default.
// Configura inoltre un interceptor per le richieste: prima di inviare ogni chiamata,
// verifica la presenza di un token di sessione e, se presente, lo inietta nell'header "Authorization"
// per gestire l'autenticazione verso il backend.
const instance = axios.create({
	baseURL: __API_URL__,
	timeout: 1000 * 5
});

instance.interceptors.request.use(
	(config) => {
		const token = sessionStorage.getItem("token");
		
		if (token) {
			config.headers["Authorization"] = `Bearer ${token}`;
		}
		
		return config;
	},
	(error) => {
		return Promise.reject(error);
	}
);

export default instance;