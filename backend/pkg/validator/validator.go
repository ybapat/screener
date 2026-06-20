package validator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/ybapat/screener/backend/pkg/apierror"
)

var validate = validator.New()

// DecodeAndValidate decodes JSON from the request body and validates struct tags.
// Returns *apierror.APIError so callers can pass it directly to response.Error.
func DecodeAndValidate(r *http.Request, dst any) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return apierror.BadRequest(fmt.Sprintf("invalid JSON: %s", err.Error()))
	}
	return ValidateStruct(dst)
}

// ValidateStruct validates a struct using struct tags.
func ValidateStruct(dst any) error {
	if err := validate.Struct(dst); err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			var msgs []string
			for _, fe := range ve {
				msgs = append(msgs, fmt.Sprintf("field '%s' %s", fe.Field(), msgForTag(fe)))
			}
			return apierror.UnprocessableEntity(strings.Join(msgs, "; "))
		}
		return apierror.UnprocessableEntity(err.Error())
	}
	return nil
}

// Validate is an alias for ValidateStruct kept for backward compatibility.
func Validate(dst any) error {
	return ValidateStruct(dst)
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
	case "gte":
		return fmt.Sprintf("must be >= %s", fe.Param())
	case "lte":
		return fmt.Sprintf("must be <= %s", fe.Param())
	case "email":
		return "must be a valid email"
	case "oneof":
		return fmt.Sprintf("must be one of [%s]", fe.Param())
	case "gtfield":
		return fmt.Sprintf("must be after %s", fe.Param())
	case "len":
		return fmt.Sprintf("must be exactly %s characters", fe.Param())
	case "alpha":
		return "must contain only letters"
	case "alphanum":
		return "must contain only letters and numbers"
	case "numeric":
		return "must be numeric"
	case "url":
		return "must be a valid URL"
	case "uri":
		return "must be a valid URI"
	case "uuid":
		return "must be a valid UUID"
	case "uuid4":
		return "must be a valid UUID v4"
	default:
		return fmt.Sprintf("failed on '%s' validation", fe.Tag())
	}
}
