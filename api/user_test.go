package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/lib/pq"
	mockdb "github.com/shimon-git/simple-bank/db/mock"
	db "github.com/shimon-git/simple-bank/db/sqlc"
	"github.com/shimon-git/simple-bank/util"
	"github.com/stretchr/testify/require"
)

/*
 * creating custom eqCreateUserParamsMatcher + receivers
 * this customization needed because each time we generating a new hash to the same password
 * this custom eqCreateUserParamsMatcher was created to handle with this 'problem'
 */

// eqCreateUserParamsMatcher -  custom type for comparing between 2 users + passwords
type eqCreateUserParamsMatcher struct {
	arg      db.CreateUserParams
	password string
}

// Matches - return true if the there is a match otherwise return false
func (e eqCreateUserParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	if err := util.CheckPassword(e.password, arg.HashedPassword); err != nil {
		return false
	}

	e.arg.HashedPassword = arg.HashedPassword

	return reflect.DeepEqual(e.arg, arg)
}

// String - return a string log
func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

// EqreateUserParams - this function responsible
// to call the eqCreateUserParamsMatcher with the given params and return Matcher  type
func EqreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{arg, password}
}

func TestCreateUser(t *testing.T) {
	user, password := randomUser(t)

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorded *httptest.ResponseRecorder)
	}{{
		name: "OK",
		body: gin.H{
			"username":  user.Username,
			"password":  password,
			"full_name": user.FullName,
			"email":     user.Email,
		},
		buildStubs: func(store *mockdb.MockStore) {
			arg := db.CreateUserParams{
				Username: user.Username,
				FullName: user.FullName,
				Email:    user.Email,
			}
			store.EXPECT().
				CreateUser(gomock.Any(), EqreateUserParams(arg, password)).
				Times(1).
				Return(user, nil)
		}, checkResponse: func(t *testing.T, recorded *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusOK, recorded.Code)
			requireBodyMatchUser(t, recorded.Body, user)
		},
	}, {
		name: "InternalError",
		body: gin.H{
			"username":  user.Username,
			"password":  password,
			"full_name": user.FullName,
			"email":     user.Email,
		},
		buildStubs: func(store *mockdb.MockStore) {
			store.EXPECT().
				CreateUser(gomock.Any(), gomock.Any()).
				Times(1).
				Return(db.User{}, sql.ErrConnDone)
		},
		checkResponse: func(t *testing.T, recorded *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusInternalServerError, recorded.Code)
		},
	}, {
		name: "DuplicateUsername",
		body: gin.H{
			"username":  user.Username,
			"password":  password,
			"full_name": user.FullName,
			"email":     user.Email,
		},
		buildStubs: func(store *mockdb.MockStore) {
			store.EXPECT().
				CreateUser(gomock.Any(), gomock.Any()).
				Times(1).
				Return(db.User{}, &pq.Error{Code: "23505"})
		},
		checkResponse: func(t *testing.T, recorded *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusForbidden, recorded.Code)
		},
	}, {
		name: "InvalidUsername",
		body: gin.H{
			"username":  "invalid-user#",
			"password":  password,
			"full_name": user.FullName,
			"email":     user.Email,
		},
		buildStubs: func(store *mockdb.MockStore) {
			store.EXPECT().
				CreateUser(gomock.Any(), gomock.Any()).
				Times(0)
		},
		checkResponse: func(t *testing.T, recorded *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusBadRequest, recorded.Code)
		},
	}, {
		name: "InvalidEmail",
		body: gin.H{
			"username":  user.Username,
			"password":  password,
			"full_name": user.FullName,
			"email":     "invalid-email",
		},
		buildStubs: func(store *mockdb.MockStore) {
			store.EXPECT().
				CreateUser(gomock.Any(), gomock.Any()).
				Times(0)
		},
		checkResponse: func(t *testing.T, recorded *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusBadRequest, recorded.Code)
		},
	}, {
		name: "TooShortPassword",
		body: gin.H{
			"username":  user.Username,
			"password":  "123",
			"full_name": user.FullName,
			"email":     "invalid-email",
		},
		buildStubs: func(store *mockdb.MockStore) {
			store.EXPECT().
				CreateUser(gomock.Any(), gomock.Any()).
				Times(0)
		},
		checkResponse: func(t *testing.T, recorded *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusBadRequest, recorded.Code)
		},
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewTestServer(t, store)
			recorder := httptest.NewRecorder()
			url := "/users"

			reqBody, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
			require.NoError(t, err)

			server.Router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(6)
	passwordHash, err := util.HashedPassword(password)
	require.NoError(t, err)

	user = db.User{
		Username:       util.RandomString(6),
		HashedPassword: passwordHash,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}
	return user, password
}

// requireBodyMatchUser - comparing the user that we request to create and the response of this request
func requireBodyMatchUser(t *testing.T, recordedBody *bytes.Buffer, user db.User) {
	var res userResponse
	err := json.NewDecoder(recordedBody).Decode(&res)
	require.NoError(t, err)

	require.Equal(t, user.Username, res.Username)
	require.Equal(t, user.FullName, res.FullName)
	require.Equal(t, user.Email, res.Email)
}
