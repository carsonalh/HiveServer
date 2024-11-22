package main

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"os"
)

const (
	jwtInvalidToken = iota + 1
	jwtSignedMalformedToken
)

type jwtParseError struct {
	code    int
	message string
}

func (err *jwtParseError) Error() string {
	return err.message
}

func parseJwt(r *http.Request) (uint64, *jwtParseError) {
	authorization := r.Header.Get("Authorization")
	var tokenString string
	_, err := fmt.Sscanf(authorization, "Bearer %s", &tokenString)

	if err != nil {
		return 0, &jwtParseError{
			code:    jwtInvalidToken,
			message: fmt.Sprintf("could not parse bearer token, %v", err),
		}
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return 0, &jwtParseError{
			code:    jwtInvalidToken,
			message: fmt.Sprintf("could not decode jwt, %v", err),
		}
	}

	if !token.Valid {
		return 0, &jwtParseError{
			code:    jwtInvalidToken,
			message: "invalid token",
		}
	}

	leakedKeyError := &jwtParseError{
		code:    jwtSignedMalformedToken,
		message: fmt.Sprintf("submitted signed jwt with an invalid id, THE SECRET KEY HAS BEEN LEAKED, token claims = %v", token.Claims),
	}

	lookupId, lookupOk := token.Claims.(jwt.MapClaims)["id"]

	if !lookupOk {
		return 0, leakedKeyError
	}

	idFloat, isFloat := lookupId.(float64)

	if !isFloat {
		return 0, leakedKeyError
	}

	id := uint64(idFloat)

	return id, nil
}

// loadPlayerId securely loads a player id from the request Authorization header.
// Handles all logging and error cases itself, just check the second return value to see if it
// succeeded.
func loadPlayerId(w http.ResponseWriter, r *http.Request) (uint64, bool) {
	playerId, tokenError := parseJwt(r)

	if tokenError != nil {
		if tokenError.code == jwtSignedMalformedToken {
			log.Printf("%s %s, %s", r.Method, r.URL.Path, tokenError.Error())
		}

		w.WriteHeader(http.StatusUnauthorized)
		return 0, false
	}

	return playerId, true
}
