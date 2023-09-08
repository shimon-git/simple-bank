package db

import (
	"context"
	"database/sql"
	"fmt"
)

// store provides all functions to execute db queries and transactions
type Store interface {
	Querier
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
}

// * Store provides all functions to execute db queries and transactions
type SQLStore struct {
	*Queries
	db *sql.DB
}

// * NewStore - create a new store
func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

/*
 * execTx - execute a function within a database transaction
 * params:
 * ctx - context
 * fn - the given function where the transaction done(multiply queries)
 */
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	// creating transaction object
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// creating new queries - transaction object
	q := New(tx)
	// passing the queries transaction params into the given function
	err = fn(q)
	// checking for errors in the transaction
	if err != nil {
		// rolling back - and checking for errors while rolling back
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %v, rollback error: %v", err, rbErr)
		}
		return fmt.Errorf("transaction error: %v", err)
	}

	// if transaction was done successfully commit the transaction and return commits errors
	return tx.Commit()
}

// * TransferTxParams - contains the input parameters of the transfer transaction
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// * TransferTxResults - contains the output of the transfer transaction
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

/*
* TransferT - preforms a money transfer from one account to the other
* I) creates a transfer record
* II) add account entries,
* III) update account's balance
* within a single database transaction
 */
func (store *SQLStore) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	// creating the result object of the transaction
	var result TransferTxResult
	/*
	* Running the transaction
	* The transaction steps are:
	* 1. creating transfer
	* 2. creating entries - (2 entries one from account and the other for to account)
	 */
	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		result.Transfer, err = makeTransfer(ctx, q, arg.FromAccountID, arg.ToAccountID, arg.Amount)
		if err != nil {
			return err
		}

		result.FromEntry, result.ToEntry, err = makeEntry(ctx, q, arg.FromAccountID, arg.ToAccountID, arg.Amount)
		if err != nil {
			return err
		}

		if arg.FromAccountID < arg.ToAccountID {
			// update balance account's
			result.FromAccount, err = updateAccountBalance(ctx, q, arg.FromAccountID, (arg.Amount * -1))
			if err != nil {
				return err
			}

			result.ToAccount, err = updateAccountBalance(ctx, q, arg.ToAccountID, arg.Amount)
			return err

		} else {
			result.ToAccount, err = updateAccountBalance(ctx, q, arg.ToAccountID, arg.Amount)
			if err != nil {
				return err
			}

			result.FromAccount, err = updateAccountBalance(ctx, q, arg.FromAccountID, (arg.Amount * -1))
			return err
		}
	})

	// returning the transaction result + thr transaction error
	return result, err
}

// makeTransfer - creating a transfer
func makeTransfer(ctx context.Context, q *Queries, fromAccountID, toAccountID, amount int64) (Transfer, error) {
	return q.CreateTransfer(ctx, CreateTransferParams{
		FromAccountID: fromAccountID,
		ToAccountID:   toAccountID,
		Amount:        amount,
	})
}

// makeTransfer - creating entries for from account + to account
func makeEntry(ctx context.Context, q *Queries, fromAccountID, toAccountID, amount int64) (Entry, Entry, error) {
	fromEntryResult, err := q.CreateEntry(ctx, CreateEntryParams{
		AccountID: fromAccountID,
		Amount:    (amount * -1),
	})
	if err != nil {
		return Entry{}, Entry{}, err
	}

	toEntryResult, err := q.CreateEntry(ctx, CreateEntryParams{
		AccountID: toAccountID,
		Amount:    amount,
	})

	return fromEntryResult, toEntryResult, err

}

// updateAccountBalance - update the account balance
func updateAccountBalance(ctx context.Context, q *Queries, accountID, addedBalance int64) (Account, error) {

	// getting the account details and locking the DB for update
	account, err := q.GetAccountForUpdate(ctx, accountID)
	if err != nil {
		return Account{}, err
	}

	// returning the update details account + errors
	return q.UpdateAccount(ctx, UpdateAccountParams{
		ID:      accountID,
		Balance: account.Balance + addedBalance,
	})
}
