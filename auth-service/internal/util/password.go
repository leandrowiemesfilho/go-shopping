package util

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

// PasswordUtil provides methods for hashing and verifying passwords
type PasswordUtil interface {
	HashPassword(password string) (string, error)
	VerifyPassword(password, hash string) bool
	PasswordStrength(password string) []string
	GenerateRandomPassword(length int) (string, error)
	ValidatePassword(password string) error
}

type passwordUtil struct{}

// NewPasswordUtil creates a new instance of PasswordUtil
func NewPasswordUtil() PasswordUtil {
	return &passwordUtil{}
}

// HashPassword hashes a plain text password using bcrypt
func (p *passwordUtil) HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	if len(password) < 6 {
		return "", fmt.Errorf("password must be at least 6 characters long")
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// VerifyPassword compares a plain text password with a bcrypt hash
func (p *passwordUtil) VerifyPassword(password, hash string) bool {
	if password == "" || hash == "" {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// PasswordStrength checks the strength of a password and returns validation errors
func (p *passwordUtil) PasswordStrength(password string) []string {
	var errors []string

	if len(password) < 8 {
		errors = append(errors, "password must be at least 8 characters long")
	}

	// Check for at least one uppercase letter
	hasUpper := false
	// Check for at least one lowercase letter
	hasLower := false
	// Check for at least one digit
	hasDigit := false
	// Check for at least one special character
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case char >= '!' && char <= '/', char >= ':' && char <= '@', char >= '[' && char <= '`', char >= '{' && char <= '~':
			hasSpecial = true
		}
	}

	if !hasUpper {
		errors = append(errors, "password must contain at least one uppercase letter")
	}
	if !hasLower {
		errors = append(errors, "password must contain at least one lowercase letter")
	}
	if !hasDigit {
		errors = append(errors, "password must contain at least one digit")
	}
	if !hasSpecial {
		errors = append(errors, "password must contain at least one special character")
	}

	return errors
}

// cryptoSafeIntn generates a cryptographically secure random integer in [0, n)
func cryptoSafeIntn(n int) (int, error) {
	if n <= 0 {
		return 0, fmt.Errorf("n must be positive")
	}

	maxValue := big.NewInt(int64(n))
	randomNum, err := rand.Int(rand.Reader, maxValue)
	if err != nil {
		return 0, err
	}

	return int(randomNum.Int64()), nil
}

// cryptoSafeShuffle shuffles a slice using cryptographically secure random numbers
func cryptoSafeShuffle(slice []byte) error {
	for i := len(slice) - 1; i > 0; i-- {
		j, err := cryptoSafeIntn(i + 1)
		if err != nil {
			return err
		}
		slice[i], slice[j] = slice[j], slice[i]
	}
	return nil
}

// GenerateRandomPassword generates a random password with specified length and complexity
func (p *passwordUtil) GenerateRandomPassword(length int) (string, error) {
	if length < 8 {
		return "", fmt.Errorf("password length must be at least 8 characters")
	}

	const (
		upperChars   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		lowerChars   = "abcdefghijklmnopqrstuvwxyz"
		digitChars   = "0123456789"
		specialChars = "!@#$%^&*()-_=+,.?/:;{}[]~"
	)

	allChars := upperChars + lowerChars + digitChars + specialChars

	// Ensure we have at least one character from each character set
	password := make([]byte, length)

	// Set first four characters to one from each character set
	if idx, err := cryptoSafeIntn(len(upperChars)); err == nil {
		password[0] = upperChars[idx]
	} else {
		return "", err
	}

	if idx, err := cryptoSafeIntn(len(lowerChars)); err == nil {
		password[1] = lowerChars[idx]
	} else {
		return "", err
	}

	if idx, err := cryptoSafeIntn(len(digitChars)); err == nil {
		password[2] = digitChars[idx]
	} else {
		return "", err
	}

	if idx, err := cryptoSafeIntn(len(specialChars)); err == nil {
		password[3] = specialChars[idx]
	} else {
		return "", err
	}

	// Fill the rest with random characters from all sets
	for i := 4; i < length; i++ {
		if idx, err := cryptoSafeIntn(len(allChars)); err == nil {
			password[i] = allChars[idx]
		} else {
			return "", err
		}
	}

	// Shuffle the password to avoid predictable pattern
	if err := cryptoSafeShuffle(password); err != nil {
		return "", err
	}

	return string(password), nil
}

// Simple random int generator for testing (not cryptographically secure)
// This should only be used in tests, not in production
func simpleIntn(n int) int {
	if n <= 0 {
		return 0
	}

	// Use a simple pseudo-random approach for testing
	var buf [8]byte
	_, _ = rand.Read(buf[:]) // This is still crypto/rand, but we're not handling errors for simplicity
	randomNum := binary.BigEndian.Uint64(buf[:])
	return int(randomNum % uint64(n))
}

// ValidatePassword performs comprehensive password validation
func (p *passwordUtil) ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	if len(password) > 72 { // bcrypt limit
		return fmt.Errorf("password cannot exceed 72 characters")
	}

	strengthErrors := p.PasswordStrength(password)
	if len(strengthErrors) > 0 {
		return fmt.Errorf("password is weak: %v", strengthErrors)
	}

	return nil
}
