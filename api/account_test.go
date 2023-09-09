package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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

func TestGetAccountAPI(t *testing.T) {
	// creating random account
	user, _ := randomUser(t)
	account := randomAccount(user.Username)

	testCases := []struct {
		name          string
		accountID     int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorded *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// building the expected function to be execute and the expected results
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check response
				log.Println(recorder.Body.String())
				require.Equal(t, http.StatusOK, recorder.Code)

				requiredBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// building the expected function to be execute and the expected results
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check response
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// building the expected function to be execute and the expected results
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check response
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			accountID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// building the expected function to be execute and the expected results
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check response
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			// creating a new mock controller
			ctrl := gomock.NewController(t)
			// check the all methods that related to the controller was finished
			ctrl.Finish()
			// creating a new mock DB controller
			store := mockdb.NewMockStore(ctrl)
			// building the stubs
			tc.buildStubs(store)
			// starting test server for sending & testing requests
			server := NewTestServer(t, store)
			// creating a new recorder
			recorder := httptest.NewRecorder()
			// specifying the url
			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			// creating a new http request & checking for errors
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)
			// adding the authorization header
			tc.setupAuth(t, req, server.token)
			// sending the request to the router and record the http response
			server.Router.ServeHTTP(recorder, req)
			// checking the response
			tc.checkResponse(t, recorder)

		})
	}
}

// randomAccount - returning random account
func randomAccount(owner string) db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    owner,
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

// requiredBodyMatchAccount - checking the given body is equal to the expected body
func requiredBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	// reading the response body from the buffer
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)
	// retrieving the response into the gotAccount interface
	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)
	// testing the response
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}
