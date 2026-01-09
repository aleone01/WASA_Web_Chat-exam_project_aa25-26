package api

import (
	"net/http"
)

// Handler configura e restituisce l'oggetto http.Handler principale che gestisce il routing delle richieste API.
// Questa funzione mappa i metodi HTTP e i percorsi URL (endpoint) alle specifiche funzioni del router (_router),
// avvolgendole con il middleware 'wrap' per iniettare il contesto della richiesta dove necessario.
func (rt *_router) Handler() http.Handler {

	// rotta di liveness
	rt.router.GET("/liveness", rt.liveness) // Definito nel tuo liveness.go

	// rotta di login
	rt.router.POST("/login", rt.wrap(rt.doLogin))

	// rotte utente
	rt.router.PUT("/users/:userId/username", rt.wrap(rt.setMyUserName))
	rt.router.PUT("/users/:userId/photo", rt.wrap(rt.setMyPhoto))

	// rotte conversazione
	rt.router.POST("/users/:userId/chats", rt.wrap(rt.createConversation))
	rt.router.GET("/users/:userId/chats", rt.wrap(rt.getMyConversations))
	rt.router.GET("/users/:userId/chats/:chatId", rt.wrap(rt.getConversation))

	// rotte messaggi
	rt.router.POST("/users/:userId/chats/:chatId/messages", rt.wrap(rt.sendMessage))
	rt.router.DELETE("/users/:userId/chats/:chatId/messages/:messageId", rt.wrap(rt.deleteMessage))
	rt.router.POST("/users/:userId/chats/:chatId/messages/:messageId/forward", rt.wrap(rt.forwardMessage))

	// rotte reazioni ai messaggi
	rt.router.POST("/users/:userId/chats/:chatId/messages/:messageId/reactions", rt.wrap(rt.commentMessage))
	rt.router.DELETE("/users/:userId/chats/:chatId/messages/:messageId/reactions", rt.wrap(rt.uncommentMessage))

	// rotte gruppi
	rt.router.POST("/groups", rt.wrap(rt.createGroup))
	rt.router.GET("/groups/:groupId/members", rt.wrap(rt.getGroupMembers))
	rt.router.POST("/groups/:groupId/members", rt.wrap(rt.addToGroup))
	rt.router.DELETE("/groups/:groupId/leave", rt.wrap(rt.leaveGroup))
	rt.router.PUT("/groups/:groupId/name", rt.wrap(rt.setGroupName))
	rt.router.PUT("/groups/:groupId/photo", rt.wrap(rt.setGroupPhoto))

	// Mappa l'URL /images/* alla cartella locale /tmp/wasa_images
	rt.router.ServeFiles("/images/*filepath", http.Dir("/tmp/wasa_images"))

	return rt.router
}
