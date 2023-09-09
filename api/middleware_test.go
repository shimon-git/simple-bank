package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shimon-git/simple-bank/token"
	"github.com/stretchr/testify/require"
)

// addAuthorization - adding the authorization header by the given params
func addAuthorization(
	t *testing.T,
	request *http.Request,
	tokenMaker token.Maker,
	authorizationType string,
	username string,
	duration time.Duration,
) {
	// creating a new token
	token, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	// creating the authorization header value
	authorizationHeader := fmt.Sprintf("%s %s", authorizationType, token)
	// setting the authorization header
	request.Header.Set(authorizationHeaderKey, authorizationHeader)
}

func TestAuthMiddleware(t *testing.T) {
	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		checkResponse func(t *testing.T, recoredr *httptest.ResponseRecorder)
	}{{
		name: "OK",
		setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "user", time.Minute)
		},
		checkResponse: func(t *testing.T, recorded *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusOK, recorded.Code)
		},
	},
		{
			name: "NotAuthorization",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// leaving this function empty because we want to test a case
				// without an authorization header
			},
			checkResponse: func(t *testing.T, recorded *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorded.Code)
			},
		},
		{
			name: "UnsupportedAuthorization",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, "unsupported-authorization", "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorded *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorded.Code)
			},
		},
		{
			name: "InvalidAuthorizationFormat",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, "", "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorded *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorded.Code)
			},
		},
		{
			name: "ExpiredToken",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "user", -time.Minute)
			},
			checkResponse: func(t *testing.T, recorded *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorded.Code)
			},
		}}

	// looping through the test cases
	for testIDX := range testCases {
		// current test
		test := testCases[testIDX]
		// running sub tests in goroutines
		t.Run(test.name, func(t *testing.T) {
			// creating a new test server
			server := NewTestServer(t, nil)
			// the uri path
			authPath := "/auth"
			// creating a new endpoint handler
			server.Router.GET(
				authPath,
				authMiddleware(server.token),
				func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{})
				},
			)
			// creating a response recorder
			recorder := httptest.NewRecorder()
			// creating a new http request
			request, err := http.NewRequest(http.MethodGet, authPath, nil)
			require.NoError(t, err)
			// setting up the auth headers
			test.setupAuth(t, request, server.token)
			// sending the request to the http handler
			server.Router.ServeHTTP(recorder, request)
			// checking the response
			test.checkResponse(t, recorder)
		})
	}
}
