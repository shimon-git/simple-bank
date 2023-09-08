package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/shimon-git/simple-bank/db/sqlc"
)

// Server serves HTTP requests for our banking service
type Server struct {
	store  db.Store
	Router *gin.Engine
}

// NewServer - creates a new HTTP server and setup routing
func NewServer(store db.Store) *Server {
	// creating a new server object
	server := &Server{
		store:  store,
		Router: gin.Default(),
	}
	// defining the router routes
	server.Router.POST("/accounts", server.createAccount)
	server.Router.GET("/accounts", server.listAccounts)
	server.Router.GET("/accounts/:id", server.getAccount)
	return server
}

// start - starting the HTTP server on a specific address
func (server *Server) Start(address string) error {
	return server.Router.Run(address)
}

// errorResponse - return map[string]{} - we inside the map[string]VALUE the error response
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
