package util


const (
	USD = "USD"
	EUR = "EUR"
	AUD = "AUD"
)

//IsSupportedCurrency returns true if the currency is supported
func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, EUR, AUD:
		return true
	}
	return false
}