package validation

import (
	"regexp"
)

// Validation rule patterns
var (
	// Email validation pattern - configurable
	EmailPattern = `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`

	// Student identifier pattern - 8 digits
	IdentifierPattern = `^\d{8}$`

	// Password min length
	PasswordMinLength = 8

	// Name validation min/max length
	NameMinLength = 2
	NameMaxLength = 100
)

// CompiledPatterns caches compiled regex patterns for better performance
var CompiledPatterns = struct {
	Email      *regexp.Regexp
	Identifier *regexp.Regexp
}{
	Email:      regexp.MustCompile(EmailPattern),
	Identifier: regexp.MustCompile(IdentifierPattern),
}

// String validation
type StringValidation struct {
	Value    string
	MinLen   int
	MaxLen   int
	Required bool
	Pattern  *regexp.Regexp
}

// NewStringValidation creates a new string validation
func NewStringValidation(value string) *StringValidation {
	return &StringValidation{
		Value:    value,
		Required: true,
	}
}

// WithMinLength sets minimum length
func (v *StringValidation) WithMinLength(min int) *StringValidation {
	v.MinLen = min
	return v
}

// WithMaxLength sets maximum length
func (v *StringValidation) WithMaxLength(max int) *StringValidation {
	v.MaxLen = max
	return v
}

// WithPattern sets regex pattern
func (v *StringValidation) WithPattern(pattern *regexp.Regexp) *StringValidation {
	v.Pattern = pattern
	return v
}

// WithRequired sets if field is required
func (v *StringValidation) WithRequired(required bool) *StringValidation {
	v.Required = required
	return v
}

// Validate performs validation
func (v *StringValidation) Validate() bool {
	// Check if required
	if v.Required && v.Value == "" {
		return false
	}

	// Skip other validations for empty optional values
	if !v.Required && v.Value == "" {
		return true
	}

	// Check min length
	if v.MinLen > 0 && len(v.Value) < v.MinLen {
		return false
	}

	// Check max length
	if v.MaxLen > 0 && len(v.Value) > v.MaxLen {
		return false
	}

	// Check pattern
	if v.Pattern != nil && !v.Pattern.MatchString(v.Value) {
		return false
	}

	return true
}

// Numeric validation
type NumericValidation struct {
	Value    int
	Min      int
	Max      int
	Required bool
}

// NewNumericValidation creates a new numeric validation
func NewNumericValidation(value int) *NumericValidation {
	return &NumericValidation{
		Value:    value,
		Required: true,
	}
}

// WithMin sets minimum value
func (v *NumericValidation) WithMin(min int) *NumericValidation {
	v.Min = min
	return v
}

// WithMax sets maximum value
func (v *NumericValidation) WithMax(max int) *NumericValidation {
	v.Max = max
	return v
}

// WithRequired sets if field is required
func (v *NumericValidation) WithRequired(required bool) *NumericValidation {
	v.Required = required
	return v
}

// Validate performs validation
func (v *NumericValidation) Validate() bool {
	// Check min value
	if v.Min != 0 && v.Value < v.Min {
		return false
	}

	// Check max value
	if v.Max != 0 && v.Value > v.Max {
		return false
	}

	return true
}
