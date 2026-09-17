package backEnd

import (
	"errors"
	"net/mail"
	"strings"
	"unicode"
	"unicode/utf8"

	
)

func ValidateDisplayName(name string) (string, error) {

	if name == "" {
		return "", errors.New("display name is required")
	}

	if !utf8.ValidString(name) {
		return "", errors.New("display name contains invalid character")
	}

	for _, r := range name {
		if unicode.IsControl(r) {
			return "", errors.New("display name contains invalid characters")
		}
	}

	name = strings.TrimSpace(name)

	name = strings.Join(strings.Fields(name), " ")

	length := utf8.RuneCountInString(name)

	if length < 2 || length > 10 {
		return "", errors.New("display name must be between 2 and 10 characters")
	}

	for _, r := range name {
		if unicode.IsLetter(r) ||
			unicode.IsNumber(r) ||
			r == ' ' {
			continue
		}

		return "", errors.New("display name contains unsupported characters")
	}

	return name, nil
}

func ValidateUsername(username string) (string, error) {

	if username == "" {
		return "", errors.New("username is required")
	}

	username = strings.TrimSpace(username)

	if username == "" {
		return "", errors.New("username is required")
	}

	length := utf8.RuneCountInString(username)

	if length < 3 || length > 10 {
		return "", errors.New("username must be between 3 and 10 characters")
	}

	for i, r := range username {

		if i == 0 {
			if !((r >= 'a' && r <= 'z') ||
				(r >= 'A' && r <= 'Z')) {
				return "", errors.New("username must start with a letter")
			}
			continue
		}

		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '_' {
			continue
		}

		return "", errors.New(
			"username can only contain letters, numbers, and underscores",
		)
	}

	return username, nil
}

func ValidateEmail(email string) (string, error) {

	email = strings.TrimSpace(email)

	if email == "" {
		return "", errors.New("email is required")
	}

	address, err := mail.ParseAddress(email)
	if err != nil {
		return "", errors.New("invalid email address")
	}

	if address.Address != email {
		return "", errors.New("invalid email address")
	}

	if len(email) > 25 {
		return "", errors.New("email address is too long")
	}

	parts := strings.Split(email, "@")

	if len(parts) != 2 {
		return "", errors.New("invalid email address")
	}

	localPart := parts[0]
	domain := strings.ToLower(parts[1])

	email = localPart + "@" + domain

	return email, nil
}

func ValidatePassword(password string) (string, error) {

	if len(password) < 8 {
		return "", errors.New("password must be at least 8 characters")
	}

	var hasUpper bool
	var hasLower bool
	var hasNumber bool
	var hasSpecial bool

	for _, r := range password {
		switch {
		case r >= 'A' && r <= 'Z':
			hasUpper = true

		case r >= 'a' && r <= 'z':
			hasLower = true

		case r >= '0' && r <= '9':
			hasNumber = true

		case strings.ContainsRune("!@#$%^&*()", r):
			hasSpecial = true

		default:
			return "", errors.New(
				"password can only contain letters, numbers, and !@#$%^&*()",
			)
		}
	}

	if !hasUpper {
		return "", errors.New("password must contain at least one uppercase letter")
	}

	if !hasLower {
		return "", errors.New("password must contain at least one lowercase letter")
	}

	if !hasNumber {
		return "", errors.New("password must contain at least one number")
	}

	if !hasSpecial {
		return "", errors.New(
			"password must contain at least one special character: !@#$%^&*()",
		)
	}

	return password, nil
}

func ValidatePhoneNumber(phone string) (string, error) {
	phone = strings.TrimSpace(phone)

	if phone == "" {
		return "", errors.New("phone number is required")
	}

	if len(phone) != 11 {
		return "", errors.New("phone number must be 11 digits")
	}

	for _, r := range phone {
		if r < '0' || r > '9' {
			return "", errors.New("phone number can only contain digits")
		}
	}

	return phone, nil
}


