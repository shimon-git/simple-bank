package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mockdb "github.com/shimon-git/simple-bank/db/mock"
	db "github.com/shimon-git/simple-bank/db/sqlc"
	"github.com/shimon-git/simple-bank/token"
	"github.com/shimon-git/simple-bank/util"
	"github.com/stretchr/testify/require"
)

func TestCreateTransferAPI(t *testing.T) {
	// creating random accounts for the transfer request
	user1, _ := randomUser(t)
	user2, _ := randomUser(t)
	account1 := randomAccount(user1.Username)
	account2 := randomAccount(user2.Username)

	// setting the currency odf the random accounts to be equal
	account1.Currency = "ILS"
	account2.Currency = "ILS"

	// building the struct slices for testing case to be execute
	testCases := []struct {
		name          string
		request       transferRequest
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorded *httptest.ResponseRecorder)
	}{{
		name: "OK",
		request: transferRequest{
			FromAccountID: account1.ID,
			ToAccountID:   account2.ID,
			Amount:        util.RandomMoney(),
			Currency:      account1.Currency,
		},
		setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, time.Minute)
		},
		buildStubs: func(store *mockdb.MockStore) {
			// building the expected function to be execute and the expected results
			store.EXPECT().
				GetAccount(gomock.Any(), gomock.Eq(account1.ID)).
				Times(1).
				Return(account1, nil)

			store.EXPECT().
				GetAccount(gomock.Any(), gomock.Eq(account2.ID)).
				Times(1).
				Return(account2, nil)

			store.EXPECT().
				TransferTx(gomock.Any(), gomock.Any()).
				Times(1).
				Return(db.TransferTxResult{
					FromAccount: account1,
					ToAccount:   account2,
				}, nil)
		},
		checkResponse: func(t *testing.T, recorded *httptest.ResponseRecorder) {
			// check the http test response
			require.Equal(t, http.StatusOK, recorded.Code)

			// extracting the response in5o the result varia
			var result db.TransferTxResult
			err := json.Unmarshal(recorded.Body.Bytes(), &result)
			require.NoError(t, err)
			// comparing the results
			require.Equal(t, account1, result.FromAccount)
			require.Equal(t, account2, result.ToAccount)

		},
	}, {
		name: "InvalidRequest",
		request: transferRequest{
			FromAccountID: 0, // invalid account ID
			ToAccountID:   account2.ID,
			Amount:        util.RandomMoney(),
			Currency:      account2.Currency,
		},
		setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, time.Minute)
		},
		buildStubs: func(store *mockdb.MockStore) {
			store.EXPECT().
				GetAccount(gomock.Any(), gomock.Eq(0)).
				Times(0)

			store.EXPECT().
				GetAccount(gomock.Any(), gomock.Eq(account2.ID)).
				Times(0)
		},
		checkResponse: func(t *testing.T, recorded *httptest.ResponseRecorder) {
			// expecting for status code 400(BadRequest)
			require.Equal(t, http.StatusBadRequest, recorded.Code)
		},
	}, {
		name: "AccountNotExist",
		request: transferRequest{
			FromAccountID: 9999,
			ToAccountID:   account2.ID,
			Amount:        util.RandomMoney(),
			Currency:      account2.Currency,
		},
		setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, time.Minute)
		},
		buildStubs: func(store *mockdb.MockStore) {
			store.EXPECT().
				GetAccount(gomock.Any(), gomock.Eq(int64(9999))).
				Times(1).
				Return(db.Account{}, sql.ErrNoRows)

			store.EXPECT().
				GetAccount(gomock.Any(), gomock.Eq(account2.ID)).
				Times(1).
				Return(account2, nil)
		},
		checkResponse: func(t *testing.T, recorded *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusNotFound, recorded.Code)
		},
	}, {
		name: "UnMatchedCurrency",
		request: transferRequest{
			FromAccountID: account1.ID,
			ToAccountID:   account2.ID,
			Amount:        util.RandomMoney(),
			Currency:      "USD",
		},
		setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, time.Minute)
		},
		buildStubs: func(store *mockdb.MockStore) {
			store.EXPECT().
				GetAccount(gomock.Any(), gomock.Eq(account1.ID)).
				Times(1).
				Return(account1, nil)

			store.EXPECT().
				GetAccount(gomock.Any(), gomock.Eq(account2.ID)).
				Times(1).
				Return(account2, nil)
		},
		checkResponse: func(t *testing.T, recorded *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusBadRequest, recorded.Code)
		},
	}}

	// looping through the test cases
	for testIDX := range testCases {
		test := testCases[testIDX]
		// run sub tests in parallel(goroutines)
		t.Run(test.name, func(t *testing.T) {
			// creating mock controller
			ctrl := gomock.NewController(t)
			// check the related methods of the controller was finished
			ctrl.Finish()
			// creating a mockdb controller
			store := mockdb.NewMockStore(ctrl)

			// building the stubs
			test.buildStubs(store)

			// building http test server for sending & testing requests
			server := NewTestServer(t, store)
			// creating a recorder
			recorder := httptest.NewRecorder()
			// specify the url path
			url := "/transfers"

			// creating transfer http request & checking for errors
			reqBody, err := json.Marshal(test.request)
			require.NoError(t, err)
			require.NotEmpty(t, reqBody)

			req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			// adding the authorization header
			test.setupAuth(t, req, server.token)
			// sending the request to the router
			server.Router.ServeHTTP(recorder, req)
			// checking the response
			test.checkResponse(t, recorder)
		})
	}

}
