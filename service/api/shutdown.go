package api

// Close termina il ciclo di vita del router API.
// Questa funzione è responsabile della chiusura ordinata di eventuali risorse aperte, come goroutine in background
// o connessioni persistenti, prima che l'applicazione termini completamente.
func (rt *_router) Close() error {
	return nil
}
