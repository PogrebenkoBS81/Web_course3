package auth

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"os"
	"strings"
)

// TokenValid - checks if given token in the request is valid.
func TokenValid(r *http.Request) error {
	token, err := verifyToken(r)

	if err != nil {
		return err
	}

	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return errors.New("token is not valid")
	}

	return nil
}

// verifyToken - checks if given token in the request is valid.
func verifyToken(r *http.Request) (*jwt.Token, error) {
	return ParseToken(extractToken(r))
}

// ParseToken - parses given JWT token, checks it with the public key.
func ParseToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		block, rest := pem.Decode([]byte(os.Getenv("TOKEN_PUBLIC")))
		// If there is an extra data after -----END PUBLIC KEY----- - throw an error.
		if len(rest) > 0 {
			return nil, fmt.Errorf("unexpected extra block length: %d", rest)
		}

		return x509.ParsePKCS1PublicKey(block.Bytes)
	})
}

// extractToken - get the token from the request body
func extractToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearerToken, " ")

	if len(strArr) == 2 {
		return strArr[1]
	}

	return ""
}

// extract - extracts data from JWT.
func extract(token *jwt.Token) (*AccessDetails, error) {
	claims, ok := token.Claims.(jwt.MapClaims)

	if ok && token.Valid {
		accessUuid, ok := claims["access_uuid"].(string)
		username, userOk := claims["user_id"].(string)

		if ok == false || userOk == false {
			return nil, errors.New("unauthorized")
		} else {
			return &AccessDetails{
				TokenUuid: accessUuid,
				Username:  username,
			}, nil
		}
	}

	return nil, errors.New("something went wrong")
}
