package util

// Constants for all supported currencies
const (
	USD = "USD"
	ILS = "ILS"
	EUR = "EUR"
	CAD = "CAD"
)

// IsSupportedCurrency - returns boolean if the given currency is supported or not
func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, ILS, EUR, CAD:
		return true
	default:
		return false
	}
}
