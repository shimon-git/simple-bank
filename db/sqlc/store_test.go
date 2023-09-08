package db

import (
	"context"
	"testing"

	"github.com/shimon-git/simple-bank/util"
	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	// creating store object
	store := NewStore(testDB)
	// crating 2 accounts for the transfer transaction
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	// run n concurrent transfer transactions
	n := int(util.RandomInt(4, 6))
	amount := util.RandomMoney()

	// crating 2 chanel's - (one for errors and the other for the results)
	errs := make(chan error)
	results := make(chan TransferTxResult)

	// creating n go routines - one for each transaction
	for i := 0; i < n; i++ {
		go func() {
			// crating a transaction
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})
			// passing the error + result to the appropriate channel
			errs <- err
			results <- result
		}()
	}

	// check results && errors for n transactions
	for i := 0; i < n; i++ {
		// getting the errors + results from the channel
		err := <-errs
		result := <-results

		require.NoError(t, err)
		require.NotEmpty(t, result)

		// check transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)

		require.Equal(t, transfer.FromAccountID, account1.ID)
		require.Equal(t, transfer.ToAccountID, account2.ID)
		require.Equal(t, transfer.Amount, amount)

		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// check entries
		fromEntry := result.FromEntry
		toEntry := result.ToEntry

		require.NotEmpty(t, fromEntry)
		require.NotEmpty(t, toEntry)

		require.Equal(t, fromEntry.AccountID, account1.ID)
		require.NotZero(t, fromEntry.Amount, (amount * -1))

		require.Equal(t, toEntry.AccountID, account2.ID)
		require.Equal(t, toEntry.Amount, amount)

		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err1 := store.GetEntry(context.Background(), fromEntry.ID)
		_, err2 := store.GetEntry(context.Background(), toEntry.ID)

		require.NoError(t, err1)
		require.NoError(t, err2)

		// check accounts
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, fromAccount.ID, account1.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, toAccount.ID, account2.ID)

		// check the account balance's
		diff1 := account1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - account2.Balance

		require.Equal(t, diff1, diff2, amount)
		require.True(t, amount > 0)
	}

	// check the final updated balance
	updateAccount1, err1 := testQueries.GetAccount(context.Background(), account1.ID)
	updateAccount2, err2 := testQueries.GetAccount(context.Background(), account2.ID)

	require.NoError(t, err1)
	require.NoError(t, err2)

	require.NotEmpty(t, updateAccount1)
	require.NotEmpty(t, updateAccount2)

	require.Equal(t, updateAccount1.Balance, account1.Balance-(amount*int64(n)))
	require.Equal(t, updateAccount2.Balance, account2.Balance+(amount*int64(n)))
}
