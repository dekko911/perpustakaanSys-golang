package hash

import "golang.org/x/crypto/bcrypt"

// hash the plain password from request input.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// checking password in db with request input password.
func CompareHashedPassword(hashedPassword string, plainTextPassword []byte) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), plainTextPassword)
	return err == nil // true
}
