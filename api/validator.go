package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/shimon-git/simple-bank/util"
)

var validCurrency validator.Func = func(fieldLevel validator.FieldLevel) bool {
	// extracting the field as an interface and covert it to a string
	currency, ok := fieldLevel.Field().Interface().(string)
	if ok {
		// return if the given currency is supported or not
		return util.IsSupportedCurrency(currency)
	}
	return false
}
