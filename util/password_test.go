package util

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestPassword(t *testing.T) {
	// creating random password
	password := RandomString(6)
	// hashing the password
	hashedPassword1, err := HashedPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword1)
	// hashing the password
	hashedPassword2, err := HashedPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword2)
	/*
	* checking the hashes of the same password is not equal
	*  each hashing needs return a different hash for the same password
	* because the salt is randomly generated
	 */
	require.NotEqual(t, hashedPassword1, hashedPassword2)

	// validating the password checker for the hash1 + password
	err = CheckPassword(password, hashedPassword1)
	require.NoError(t, err)
	// validating the password checker for the hash2 + password
	err = CheckPassword(password, hashedPassword2)
	require.NoError(t, err)

	// validating the wrong password return error
	wrongPassword := RandomString(6)
	err = CheckPassword(wrongPassword, hashedPassword1)
	require.EqualError(t, bcrypt.ErrMismatchedHashAndPassword, err.Error())

}
