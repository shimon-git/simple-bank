package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/shimon-git/simple-bank/db/sqlc"
	"github.com/shimon-git/simple-bank/token"
	"github.com/shimon-git/simple-bank/util"
)

// Server serves HTTP requests for our banking service
type Server struct {
	config util.Config
	store  db.Store
	token  token.Maker
	Router *gin.Engine
}

// NewServer - creates a new HTTP server and setup routing
func NewServer(config util.Config, store db.Store) (*Server, error) {

	token, err := token.CreateNewToken(config.TokenType, config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	// creating a new server object
	server := &Server{
		config: config,
		store:  store,
		Router: gin.Default(),
		token:  token,
	}

	// registering a validator function named currency
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}
	// setting up the routs
	server.setupRouter()

	return server, nil
}

// setupRouter - setup the routes
func (server *Server) setupRouter() {
	server.Router.POST("/users", server.createUser)
	server.Router.POST("/users/login", server.loginUser)

	// creating a new group and using the authMiddleWare
	authRoutes := server.Router.Group("/").Use(authMiddleware(server.token))

	authRoutes.POST("/accounts", server.createAccount)
	authRoutes.GET("/accounts", server.listAccounts)
	authRoutes.GET("/accounts/:id", server.getAccount)

	authRoutes.POST("/transfers", server.createTransfer)
}

// start - starting the HTTP server on a specific address
func (server *Server) Start(address string) error {
	return server.Router.Run(address)
}

// errorResponse - return map[string]{} - we inside the map[string]VALUE the error response
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
