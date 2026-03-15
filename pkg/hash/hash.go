package hash

import "golang.org/x/crypto/bcrypt"

// hash the plain password from request input.
func MakePasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// checking password in db with the request password.
func CompareHashedPassword(hashedPassword string, plainTextPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainTextPassword))
	return err == nil // true
}
