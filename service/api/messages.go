package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/aleone01/Web-Project-repo/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// sendMessage gestisce l'invio di un nuovo messaggio in una chat specifica.
// Gestisce anche la logica di inoltro (forwarding) copiando il contenuto da un messaggio esistente.
func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Recupero dei parametri dall'URL (mittente e chat di destinazione)
	userId, _ := strconv.Atoi(ps.ByName("userId"))
	chatId, _ := strconv.Atoi(ps.ByName("chatId"))

	// Verifica che l'utente sia autenticato e corrisponda al mittente dichiarato
	if !rt.checkAuth(w, r, userId) {
		return
	}

	// Parsing del form multipart (limite max 10 MB per gestire upload di foto)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "Errore nei dati del form (assicurati di usare multipart/form-data e rispettare i limiti)")
		return
	}

	// Recupero campi testuali semplici dal form
	text := r.FormValue("text")
	replyTo, _ := strconv.Atoi(r.FormValue("replyTo"))          // ID del messaggio a cui si risponde o che si inoltra
	isForward, _ := strconv.ParseBool(r.FormValue("isForward")) // Flag booleano per indicare se è un inoltro

	// Gestione Foto (opzionale)
	var photoFile []byte

	// Tentativo di recuperare il file dal form con chiave "file"
	file, _, err := r.FormFile("file")
	if err == nil {

		// Se il file è presente, lo legge
		defer file.Close()
		photoFile, err = io.ReadAll(file)
		if err != nil {
			rt.sendError(w, http.StatusInternalServerError, 500, "Errore durante la lettura del file allegato")
			return
		}
	} else if !errors.Is(err, http.ErrMissingFile) {
		// Se l'errore non è "File mancante" (che è accettabile), allora è un errore reale
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore generico nel recupero del form file")
		return
	}

	// LOGICA DI INOLTRO
	// Se il messaggio è un inoltro, dobbiamo recuperare il contenuto del messaggio originale (replyTo)
	if isForward && replyTo > 0 {
		originalMsg, err := rt.db.GetMessage(replyTo)
		if err == nil {
			// Se il nuovo testo è vuoto, ereditiamo quello originale
			if text == "" {
				text = originalMsg.Text
			}
			// Se non c'è una nuova foto, ereditiamo quella originale
			if len(photoFile) == 0 {
				photoFile = originalMsg.PhotoFile
			}
		} else {
			rt.sendError(w, http.StatusNotFound, 404, "Impossibile trovare il messaggio originale da inoltrare")
			return
		}
	}

	// Validazione finale: il messaggio non può essere completamente vuoto (né testo né foto)
	if text == "" && len(photoFile) == 0 {
		rt.sendError(w, http.StatusBadRequest, 400, "Il messaggio non può essere vuoto (richiesto testo o foto)")
		return
	}

	sentAt := time.Now()

	// Salvataggio del messaggio nel database
	msg, err := rt.db.CreateMessage(chatId, userId, text, photoFile, sentAt, replyTo, isForward)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore nel salvataggio del messaggio su DB")
		return
	}

	// Risposta successo
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(msg)
}

// commentMessage aggiunge una reazione (emoticon) a un messaggio specifico.
func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Recupero parametri
	userId, _ := strconv.Atoi(ps.ByName("userId"))
	messageId, _ := strconv.Atoi(ps.ByName("messageId"))

	// Verifica autorizzazione: solo l'utente che commenta può aggiungere la reazione
	if !rt.checkAuth(w, r, userId) {
		return
	}

	// Parsing del JSON per ottenere l'emoticon
	var reqBody struct {
		Emoticon string `json:"emoticon"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido")
		return
	}

	// Salvataggio della reazione nel DB
	if err := rt.db.AddReaction(messageId, userId, reqBody.Emoticon); err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore durante l'aggiunta della reazione")
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// uncommentMessage rimuove una reazione precedentemente aggiunta da un utente a un messaggio.
func (rt *_router) uncommentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Recupero parametri
	userId, _ := strconv.Atoi(ps.ByName("userId"))
	messageId, _ := strconv.Atoi(ps.ByName("messageId"))

	// Verifica autorizzazione: solo l'utente che ha aggiunto la reazione può rimuoverla
	if !rt.checkAuth(w, r, userId) {
		return
	}

	// Rimozione della reazione dal DB
	if err := rt.db.RemoveReaction(messageId, userId); err != nil {
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore nella rimozione della reazione")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// deleteMessage cancella un messaggio inviato.
// Implementa un controllo per assicurare che solo l'autore possa cancellare il proprio messaggio.
func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Recupero parametri
	userId, _ := strconv.Atoi(ps.ByName("userId"))
	messageId, _ := strconv.Atoi(ps.ByName("messageId"))

	// Verifica autorizzazione: solo l'autore può cancellare il messaggio
	if !rt.checkAuth(w, r, userId) {
		return
	}

	// Tentativo di cancellazione su DB
	if err := rt.db.DeleteMessage(messageId, userId); err != nil {

		// Controllo specifico dell'errore per distinguere tra "errore server" e "permesso negato/non trovato"
		if err.Error() == "message not found or unauthorized" {
			rt.sendError(w, http.StatusForbidden, 403, "Non autorizzato a cancellare questo messaggio o messaggio inesistente")
			return
		}
		rt.sendError(w, http.StatusInternalServerError, 500, "Errore interno nella cancellazione")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// forwardMessage gestisce l'inoltro di un messaggio verso molteplici destinatari (broadcast forwarding).
// Accetta una lista di ID target (chat o gruppi) e duplica il messaggio originale per ciascuno di essi.
func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	userId, _ := strconv.Atoi(ps.ByName("userId"))
	messageId, _ := strconv.Atoi(ps.ByName("messageId"))

	if !rt.checkAuth(w, r, userId) {
		return
	}

	// Struttura body: array di interi rappresentanti gli ID delle chat/gruppi target
	var reqBody struct {
		Targets []int `json:"targets"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		rt.sendError(w, http.StatusBadRequest, 400, "JSON non valido (atteso array 'targets')")
		return
	}

	// Recupero del messaggio originale dal database
	originalMsg, err := rt.db.GetMessage(messageId)
	if err != nil {
		rt.sendError(w, http.StatusNotFound, 404, "Messaggio originale non trovato")
		return
	}

	forwardTime := time.Now()

	// Iterazione sui target per creare una copia del messaggio per ciascuna destinazione
	for _, tid := range reqBody.Targets {
		// replyTo è 0 perché è un nuovo messaggio nella nuova chat, isForward è true
		_, _ = rt.db.CreateMessage(tid, userId, originalMsg.Text, originalMsg.PhotoFile, forwardTime, 0, true)
	}

	// Restituisce il messaggio originale come conferma
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(originalMsg)
}
