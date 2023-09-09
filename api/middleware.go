package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shimon-git/simple-bank/token"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

// authMiddleware - a middleware that check for access tokens and verifying them
func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// getting the authorization header
		authorizationHeader := ctx.GetHeader("authorization")
		// if the authorization header those not exist then return error message & code 401(Unauthorized)
		if len(authorizationHeader) == 0 {
			err := errors.New("authorization header has not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		// split the authorization string into a slice
		fields := strings.Fields(authorizationHeader)
		// if the length of the authorization slice is less the 2 return error + code 401(Unauthorized)
		// 2 fields must be: 1 for the authorization type, the second for the access token value
		if len(fields) < 2 {
			err := errors.New("invalid authorization header format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		// verifying the authorization type
		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			err := fmt.Errorf("unsupported authorization type: %s", authorizationType)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		// verifying the access token
		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		// adding a new header,value for the token payload
		ctx.Set(authorizationPayloadKey, payload)
		// passing the request to the next handler
		ctx.Next()
	}
}
