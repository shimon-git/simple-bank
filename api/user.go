package api

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	db "github.com/shimon-git/simple-bank/db/sqlc"
	"github.com/shimon-git/simple-bank/util"
)

/*
* createAccountRequest -  type for creating a new account
* 'binding': validator fields - build in the gin framework
* 'oneof': validator input - the given input must be one of the 'oneof' values
 */
type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type userResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

// newUserResponse - create a new user response
func newUserResponse(user db.User) userResponse {
	return userResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
}

// createAccount - API endpoint for creating a new bank account
func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	// if one of the required fields is missed - then return an error response
	// also extracting the requests into the req variable
	// if an error ocurred return code 400(BadRequest)
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := util.HashedPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	}

	// creating an account object for recording in the DB
	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
		Email:          req.Email,
	}

	// inserting the account into the accounts table and checking for errors
	// if something goes wrong return code 500(InternalServerError)
	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		// checking if the error related to a foreign key or unique error in DB - code 403(StatusForbidden)
		if pqerr, ok := err.(*pq.Error); ok {
			switch pqerr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// creating a response
	rsp := newUserResponse(user)

	// if all good returning the new account and status OK
	ctx.JSON(http.StatusOK, rsp)
}

// loginUserRequest - a type for user login request
type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

// loginUserResponse - a type for user login response
type loginUserResponse struct {
	AccessToken string       `json:"access_token"`
	User        userResponse `json:"user"`
}

// loginUser - API endpoint for users login
func (server *Server) loginUser(ctx *gin.Context) {
	// extracting the request into the variable
	var req loginUserRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// querying the DB for the desired username
	user, err := server.store.GetUser(ctx, req.Username)
	// checking for errors
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	// verifying the password is correct
	if err = util.CheckPassword(req.Password, user.HashedPassword); err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	// generating an access token
	accessToken, err := server.token.CreateToken(user.Username, server.config.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	// creating the response
	res := loginUserResponse{
		AccessToken: accessToken,
		User:        newUserResponse(user),
	}
	// sending the response
	ctx.JSON(http.StatusOK, res)
}
