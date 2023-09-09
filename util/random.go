package util

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

// * return the seed of the given number - for ensuring we always get a random number
func init() {
	// we give as param the current unix time in nano secondes format
	rand.Seed(time.Now().UnixNano())
}

// * RandomInt generates a random integer between min and max
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// * RandomString - generates a random string of length n
func RandomString(n int) string {
	// initializing a builder string object
	var sb strings.Builder
	// calculating the alphabet length
	alphabetLength := len(alphabet)

	// iterating given n times - each iteration appending a random char to the string builder
	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(alphabetLength)]
		sb.WriteByte(c)
	}

	return sb.String()
}

// RandomOwner - generates a random owner name
func RandomOwner() string {
	return fmt.Sprintf("%s %s", RandomString(6), RandomString(6))
}

// * RandomMoney - generates a random money amount
func RandomMoney() int64 {
	return RandomInt(0, 10000)
}

// * RandomCurrency - generate a currency
func RandomCurrency() string {
	// creating a currency slice
	currencies := []string{"EUR", "USD", "ILS", "CAD"}
	// calculating the slice length
	n := len(currencies)

	return currencies[rand.Intn(n)]
}

// RandomEmail - generated random email
func RandomEmail() string {
	return fmt.Sprintf("%s@email.com", RandomString(int(RandomInt(4, 8))))
}

// RandomToken - return random token type
func RandomTokenType() string {
	suuportedTokens := []string{"JWT", "PASETO"}

	return suuportedTokens[RandomInt(0, int64(len(suuportedTokens)-1))]
}
