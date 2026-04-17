package validator

import (
	"testing"
	"time"
)

func TestValidateRequired(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     string
		wantError bool
	}{
		{"valid", "username", "john", false},
		{"empty", "username", "", true},
		{"whitespace", "username", "   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequired(tt.field, tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateRequired() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateMinLength(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     string
		min       int
		wantError bool
	}{
		{"valid", "password", "password123", 8, false},
		{"exact", "password", "12345678", 8, false},
		{"too short", "password", "pass", 8, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMinLength(tt.field, tt.value, tt.min)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateMinLength() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateMaxLength(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     string
		max       int
		wantError bool
	}{
		{"valid", "username", "john", 50, false},
		{"exact", "username", "12345", 5, false},
		{"too long", "username", "verylongusername", 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMaxLength(tt.field, tt.value, tt.max)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateMaxLength() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     string
		wantError bool
	}{
		{"valid", "email", "user@example.com", false},
		{"valid with plus", "email", "user+tag@example.com", false},
		{"no @", "email", "userexample.com", true},
		{"no domain", "email", "user@", true},
		{"no local", "email", "@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.field, tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateEmail() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateAlphanumeric(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     string
		wantError bool
	}{
		{"valid", "code", "abc123", false},
		{"letters only", "code", "abcdef", false},
		{"numbers only", "code", "123456", false},
		{"with spaces", "code", "abc 123", true},
		{"with special chars", "code", "abc-123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAlphanumeric(tt.field, tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateAlphanumeric() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateSlug(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     string
		wantError bool
	}{
		{"valid", "slug", "my-blog-post", false},
		{"single word", "slug", "post", false},
		{"with numbers", "slug", "post-123", false},
		{"uppercase", "slug", "My-Post", true},
		{"underscore", "slug", "my_post", true},
		{"trailing hyphen", "slug", "post-", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSlug(tt.field, tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateSlug() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateOneOf(t *testing.T) {
	allowed := []string{"admin", "user", "guest"}

	tests := []struct {
		name      string
		field     string
		value     string
		wantError bool
	}{
		{"valid admin", "role", "admin", false},
		{"valid user", "role", "user", false},
		{"invalid", "role", "superuser", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOneOf(tt.field, tt.value, allowed)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateOneOf() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     string
		wantError bool
	}{
		{"valid", "password", "Password123", false},
		{"too short", "password", "Pass1", true},
		{"no uppercase", "password", "password123", true},
		{"no lowercase", "password", "PASSWORD123", true},
		{"no digit", "password", "PasswordABC", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.field, tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidatePassword() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateRange(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     int
		min       int
		max       int
		wantError bool
	}{
		{"valid", "age", 25, 0, 100, false},
		{"at min", "age", 0, 0, 100, false},
		{"at max", "age", 100, 0, 100, false},
		{"below min", "age", -1, 0, 100, true},
		{"above max", "age", 101, 0, 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRange(tt.field, tt.value, tt.min, tt.max)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateRange() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidatePositive(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     int
		wantError bool
	}{
		{"positive", "count", 10, false},
		{"zero", "count", 0, true},
		{"negative", "count", -5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePositive(tt.field, tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidatePositive() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateNonNegative(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     int
		wantError bool
	}{
		{"positive", "count", 10, false},
		{"zero", "count", 0, false},
		{"negative", "count", -5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNonNegative(tt.field, tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateNonNegative() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateNotFuture(t *testing.T) {
	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	tests := []struct {
		name      string
		field     string
		value     time.Time
		wantError bool
	}{
		{"past", "timestamp", past, false},
		{"now", "timestamp", now, false},
		{"future", "timestamp", future, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNotFuture(tt.field, tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateNotFuture() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateAfter(t *testing.T) {
	start := time.Now()
	before := start.Add(-1 * time.Hour)
	after := start.Add(1 * time.Hour)

	tests := []struct {
		name      string
		field     string
		value     time.Time
		after     time.Time
		wantError bool
	}{
		{"after", "end_time", after, start, false},
		{"same", "end_time", start, start, true},
		{"before", "end_time", before, start, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAfter(tt.field, tt.value, tt.after)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateAfter() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateDuration(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		duration  time.Duration
		min       time.Duration
		max       time.Duration
		wantError bool
	}{
		{"valid", "duration", 5 * time.Minute, 1 * time.Minute, 10 * time.Minute, false},
		{"at min", "duration", 1 * time.Minute, 1 * time.Minute, 10 * time.Minute, false},
		{"at max", "duration", 10 * time.Minute, 1 * time.Minute, 10 * time.Minute, false},
		{"too short", "duration", 30 * time.Second, 1 * time.Minute, 10 * time.Minute, true},
		{"too long", "duration", 15 * time.Minute, 1 * time.Minute, 10 * time.Minute, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDuration(tt.field, tt.duration, tt.min, tt.max)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateDuration() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateSliceLength(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     []int
		min       int
		max       int
		wantError bool
	}{
		{"valid", "items", []int{1, 2, 3}, 1, 5, false},
		{"at min", "items", []int{1}, 1, 5, false},
		{"at max", "items", []int{1, 2, 3, 4, 5}, 1, 5, false},
		{"too few", "items", []int{}, 1, 5, true},
		{"too many", "items", []int{1, 2, 3, 4, 5, 6}, 1, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSliceLength(tt.field, tt.value, tt.min, tt.max)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateSliceLength() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateUnique(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     []string
		wantError bool
	}{
		{"unique", "tags", []string{"a", "b", "c"}, false},
		{"duplicate", "tags", []string{"a", "b", "a"}, true},
		{"empty", "tags", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUnique(tt.field, tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateUnique() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestCollectErrors(t *testing.T) {
	err1 := &ValidationError{Field: "field1", Message: "error1"}
	err2 := &ValidationError{Field: "field2", Message: "error2"}

	tests := []struct {
		name   string
		errors []*ValidationError
		want   int
	}{
		{"no errors", []*ValidationError{nil, nil}, 0},
		{"one error", []*ValidationError{err1, nil}, 1},
		{"two errors", []*ValidationError{err1, err2}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CollectErrors(tt.errors...)
			if len(result) != tt.want {
				t.Errorf("CollectErrors() got %d errors, want %d", len(result), tt.want)
			}
		})
	}
}

func TestValidationErrorsIsEmpty(t *testing.T) {
	tests := []struct {
		name   string
		errors ValidationErrors
		want   bool
	}{
		{"empty", ValidationErrors{}, true},
		{"not empty", ValidationErrors{{Field: "test", Message: "error"}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.errors.IsEmpty(); got != tt.want {
				t.Errorf("ValidationErrors.IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}
