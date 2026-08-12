package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Custom validation functions for business rules
type CustomValidator struct {
	validator *validator.Validate
}

func NewCustomValidator() *CustomValidator {
	v := validator.New()

	// Register custom validation functions
	v.RegisterValidation("complexpassword", validateComplexPassword)
	v.RegisterValidation("publicemail", validatePublicEmail)
	v.RegisterValidation("uuid", validateUUID)
	v.RegisterValidation("phonenumber", validatePhoneNumber)

	return &CustomValidator{validator: v}
}

func (cv *CustomValidator) ValidateStruct(s interface{}) error {
	if err := cv.validator.Struct(s); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	return nil
}

func (cv *CustomValidator) ValidateVar(field interface{}, tag string) error {
	if err := cv.validator.Var(field, tag); err != nil {
		return fmt.Errorf("field validation failed: %w", err)
	}
	return nil
}

// Custom validation functions

func validateComplexPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// At least 12 characters (as per spec)
	if len(password) < 12 {
		return false
	}

	// At least one uppercase letter
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUpper {
		return false
	}

	// At least one lowercase letter
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	if !hasLower {
		return false
	}

	// At least one digit
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasDigit {
		return false
	}

	// At least one special character
	hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)
	if !hasSpecial {
		return false
	}

	return true
}

func validatePublicEmail(fl validator.FieldLevel) bool {
	email := strings.ToLower(fl.Field().String())

	// List of public email domains that should be blocked for org registration
	publicDomains := []string{
		"gmail.com",
		"yahoo.com",
		"hotmail.com",
		"outlook.com",
		"icloud.com",
		"aol.com",
		"mail.com",
		"protonmail.com",
		"tutanota.com",
		"zoho.com",
		"yandex.com",
		"qq.com",
		"163.com",
		"sina.com",
		"sohu.com",
		"126.com",
		"yeah.net",
		"foxmail.com",
		"mail.ru",
		"web.de",
		"gmx.de",
		"libero.it",
		"virgilio.it",
		"terra.com.br",
		"uol.com.br",
		"bol.com.br",
		"ig.com.br",
		"globo.com",
		"r7.com",
	}

	// Extract domain from email
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false // Invalid email format
	}

	domain := parts[1]

	// Check if domain is in the public domains list
	for _, publicDomain := range publicDomains {
		if domain == publicDomain {
			return false // Public email domain found
		}
	}

	return true // Not a public email domain
}

func validateUUID(fl validator.FieldLevel) bool {
	uuid := fl.Field().String()

	// UUID v4 pattern: 8-4-4-4-12 hexadecimal digits
	pattern := `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-4[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`
	matched, err := regexp.MatchString(pattern, uuid)
	if err != nil {
		return false
	}

	return matched
}

func validatePhoneNumber(fl validator.FieldLevel) bool {
	phone := fl.Field().String()

	// Remove all non-digit characters
	digits := regexp.MustCompile(`[^\d]`).ReplaceAllString(phone, "")

	// Check if it has between 10 and 15 digits (typical phone number range)
	if len(digits) < 10 || len(digits) > 15 {
		return false
	}

	// Check if it starts with + (international format)
	if !strings.HasPrefix(phone, "+") {
		return false
	}

	return true
}

// Validation error messages
func GetValidationErrorMessage(err error) string {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var messages []string
		for _, e := range validationErrors {
			switch e.Tag() {
			case "required":
				messages = append(messages, fmt.Sprintf("%s is required", e.Field()))
			case "email":
				messages = append(messages, fmt.Sprintf("%s must be a valid email address", e.Field()))
			case "min":
				messages = append(messages, fmt.Sprintf("%s must be at least %s characters", e.Field(), e.Param()))
			case "max":
				messages = append(messages, fmt.Sprintf("%s must be at most %s characters", e.Field(), e.Param()))
			case "complexpassword":
				messages = append(messages, fmt.Sprintf("%s must be at least 12 characters and include uppercase, lowercase, digit, and special character", e.Field()))
			case "publicemail":
				messages = append(messages, fmt.Sprintf("%s cannot be from a public email provider (gmail, yahoo, etc.)", e.Field()))
			case "uuid":
				messages = append(messages, fmt.Sprintf("%s must be a valid UUID", e.Field()))
			case "phonenumber":
				messages = append(messages, fmt.Sprintf("%s must be a valid phone number in international format (+country number)", e.Field()))
			default:
				messages = append(messages, fmt.Sprintf("%s is invalid", e.Field()))
			}
		}
		return strings.Join(messages, "; ")
	}

	return err.Error()
}
