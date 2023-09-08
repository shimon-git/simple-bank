package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/shimon-git/simple-bank/db/sqlc"
)

/*
* createAccountRequest -  type for creating a new account
* 'binding': validator fields - build in the gin framework
* 'oneof': validator input - the given input must be one of the 'oneof' values
 */
type createAccountRequest struct {
	Owner    string `json:"owner" binding:"required"`
	Currency string `json:"currency" binding:"required,currency"`
}

// createAccount - API endpoint for creating a new bank account
func (server *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest
	// if one of the required fields is missed - then return an error response
	// also extracting the requests into the req variable
	// if an error ocurred return code 400(BadRequest)
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	// creating an account object for recording in the DB
	arg := db.CreateAccountParams{
		Owner:    req.Owner,
		Currency: req.Currency,
		Balance:  0,
	}

	// inserting the account into the accounts table and checking for errors
	// if something goes wrong return code 500(InternalServerError)
	account, err := server.store.CreateAccount(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// if all good returning the new account and status OK
	ctx.JSON(http.StatusOK, account)
}

/*
* getAccountRequest - type for getting the id account
* ID - required , minimum value for id is 1
 */
type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// getAccount - API endpoint for getting account based ID
func (server *Server) getAccount(ctx *gin.Context) {
	var req getAccountRequest
	// destructing the json request into the req variable
	// if an error ocurred return code 400(BadRequest)
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// getting the account from the DB based on the account ID
	// if an error ocurred - if sql.ErrNoRows: 400(NotFound) else 500(InternalServerError)
	account, err := server.store.GetAccount(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// if all good returning the account, status OK
	ctx.JSON(http.StatusOK, account)

}

/*
 * listAccountsRequest - type for extracting the listAccount request
 * PageID: ID to start from
 * PageSize: desired amount of rows
 */
type listAccountsRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

// listAccounts - API endpoint for listing the accounts
func (server *Server) listAccounts(ctx *gin.Context) {
	// creating a new listAccountsRequest for extracting the request
	var req listAccountsRequest
	// validating the request params - on error: status 400(BadRequest)
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	// query the DB to list the desired accounts
	accounts, err := server.store.ListAccounts(ctx, db.ListAccountsParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	})
	// checking for errors
	if err != nil {
		// if db query response with error of not rows: status 404(StatusNotFound)
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
		}
		// if the db query response with error: 500(StatusInternalServerError)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	// if all good return the accounts - status: 200(OK)
	ctx.JSON(http.StatusOK, accounts)

}
