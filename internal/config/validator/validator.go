package valid

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var v *validator.Validate

func init() {
	v = validator.New()

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		// skip if tag key says it should be ignored
		if name == "-" {
			return ""
		}
		return name
	})
}

func GetValidator() *validator.Validate {
	return v
}
