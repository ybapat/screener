package validator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func DecodeAndValidate(r *http.Request, dst any) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	if err := validate.Struct(dst); err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			var msgs []string
			for _, fe := range ve {
				msgs = append(msgs, fmt.Sprintf("field '%s' %s", fe.Field(), msgForTag(fe)))
			}
			return fmt.Errorf("%s", strings.Join(msgs, "; "))
		}
		return err
	}
	return nil
}

func msgForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "max":
		return fmt.Sprintf("must be at most %s", fe.Param())
	case "min":
		return fmt.Sprintf("must be at least %s", fe.Param())
	case "gt":
		return fmt.Sprintf("must be greater than %s", fe.Param())
	case "lt":
		return fmt.Sprintf("must be less than %s", fe.Param())
	case "email":
		return "must be a valid email"
	case "oneof":
		return fmt.Sprintf("must be one of [%s]", fe.Param())
	case "gtfield":
		return fmt.Sprintf("must be after %s", fe.Param())
	default:
		return fmt.Sprintf("failed on '%s' validation", fe.Tag())
	}
}
