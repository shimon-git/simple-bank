package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/shimon-git/simple-bank/db/sqlc"
	"github.com/shimon-git/simple-bank/token"
)

/*
* transferRequest -  type for creating a new account
* 'binding': validator fields - build in the gin framework
* 'oneof': validator input - the given input must be one of the 'oneof' values
 */
type transferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"`
}

// createAccount - API endpoint for creating a new bank account
func (server *Server) createTransfer(ctx *gin.Context) {
	var req transferRequest
	// if one of the required fields is missed - then return an error response
	// also extracting the requests into the req variable
	// if an error ocurred return code 400(BadRequest)
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// getting the owner through the user name of the access token
	authPayload, ok := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if !ok {
		err := errors.New("failed to retrieve the payload data from the authorization token")
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// validating the from account id + currency
	fromAccount, valid := server.validAccount(ctx, req.FromAccountID, req.Currency)
	if !valid {
		return
	}

	// checking if the authenticated account is authorized to make the transfer
	if fromAccount.Owner != authPayload.Username {
		err := errors.New("you are not authorized to make the transfer")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	// validating the to account id + currency
	_, valid = server.validAccount(ctx, req.ToAccountID, req.Currency)
	if !valid {
		return
	}

	// creating an account object for recording in the DB
	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	// inserting the account into the accounts table and checking for errors
	// if something goes wrong return code 500(InternalServerError)
	result, err := server.store.TransferTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// if all good returning the new account and status OK
	ctx.JSON(http.StatusOK, result)
}

// validAccount - validating the given account id + currency
func (server *Server) validAccount(ctx *gin.Context, accountID int64, currency string) (db.Account, bool) {
	// get the account by his details
	account, err := server.store.GetAccount(ctx, accountID)
	// if an error ocurred while trying to get the account
	if err != nil {
		// if the error was ocurred because the account wasn't found - status 400(StatusNotFound)
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return account, false
		}
		// if an error ocurred and it's not because the account wasn't found - status 500(StatusInternalServerError)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return account, false
	}
	// if the account currency is unmatched to the given currency - status 400(StatusBadRequest)
	if account.Currency != currency {
		err = fmt.Errorf("account [%d] currency mismatch: given currency is %s but expected currency is %s", accountID, currency, account.Currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return account, false
	}
	// if the account is valid return true
	return account, true
}
