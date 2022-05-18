package server

import (
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"log"
	"math/big"
	"net/http"
	"nexus-pusher/internal/config"
	"time"
)

type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func (u *webService) signInMiddle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var credentials Credentials
		// Get credentials from request
		user, pass, ok := r.BasicAuth()
		if !ok {
			log.Printf("error: no basic auth found in request")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		credentials.Username = user
		credentials.Password = pass

		// Get the expected password from our in memory map
		if expectedPassword, ok := u.cfg.Credentials[credentials.Username]; !ok || expectedPassword != credentials.Password {
			log.Printf("error: wrong password provided for username '%s'", credentials.Username)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Declare the expiration time of the token as 5 minutes
		expirationTime := time.Now().Add(config.JWTTokenTTL * time.Minute)
		// Create the JWT claims, which includes the username and expiry time
		claims := &Claims{
			Username: credentials.Username,
			StandardClaims: jwt.StandardClaims{
				// In JWT, the expiry time is expressed as unix milliseconds
				ExpiresAt: expirationTime.Unix(),
			},
		}

		// Declare the token with the algorithm used for signing, and the claims
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		// Create the JWT string
		tokenString, err := token.SignedString(u.jwtKey)
		if err != nil {
			// If there is an error in creating the JWT return an internal server error
			log.Printf("error: unable to create JWT token for user '%s'", credentials.Username)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Set the client cookie for "token" as the JWT we just generated
		// also set an expiry time which is the same as the token itself
		http.SetCookie(w, &http.Cookie{
			Name:    config.JWTCookieName,
			Value:   tokenString,
			Expires: expirationTime,
		})
		// Serve original request
		next.ServeHTTP(w, r)
	})
}

func (u *webService) authMiddle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to auth with client Cookie
		if _, err := u.authWithCookie(w, r); err != nil {
			return
		}
		// Serve original request
		next.ServeHTTP(w, r)
	})
}

func (u *webService) refreshMiddle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to auth with client Cookie
		claims, err := u.authWithCookie(w, r)
		if err != nil {
			return
		}

		// A new token will only be issued if the old token is within
		// 30 seconds of expiry. Otherwise, return a bad request status
		if time.Until(time.Unix(claims.ExpiresAt, 0)) > config.JWTTokenRefreshWindow*time.Second {
			log.Printf("error: token is too new to refresh. still valid for: '%v'",
				time.Until(time.Unix(claims.ExpiresAt, 0)).Round(time.Second))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// Create a new token for the current use, with a renewed expiration time
		expirationTime := time.Now().Add(config.JWTTokenTTL * time.Minute)
		claims.ExpiresAt = expirationTime.Unix()
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(u.jwtKey)
		if err != nil {
			// If there is an error in creating the JWT return an internal server error
			log.Printf("error: unable to create JWT token for user '%s'", claims.Username)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Set the new user Cookie token
		http.SetCookie(w, &http.Cookie{
			Name:    config.JWTCookieName,
			Value:   tokenString,
			Expires: expirationTime,
		})
		// Serve original request
		next.ServeHTTP(w, r)
	})
}

func genRandomJWTKey(n int) []byte {
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, n)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letter))))
		if err != nil {
			log.Fatalf("error: can't generate jwt key - %v", err)
		}
		b[i] = letter[n.Uint64()]
	}
	return []byte(string(b))
}

func (u *webService) authWithCookie(w http.ResponseWriter, r *http.Request) (*Claims, error) {
	// Get the session token from the requests cookies
	c, err := r.Cookie(config.JWTCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			// If the cookie is not set, return an unauthorized status
			log.Printf("error: cookie with jwt token is not set for this request")
			w.WriteHeader(http.StatusUnauthorized)
			return nil, err
		}
		// For any other type of error, return a bad request status
		log.Printf("error: can't get cookie for this request. %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return nil, err
	}

	// Get the JWT string from the cookie
	tknStr := c.Value

	// Initialize a new instance of `Claims`
	claims := &Claims{}

	// Parse the JWT string and store the result in `claims`.
	tkn, err := jwt.ParseWithClaims(tknStr, claims,
		func(token *jwt.Token) (interface{}, error) {
			return u.jwtKey, nil
		})
	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			log.Printf("error: can't parse JWT string. %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return nil, err
		}
		log.Printf("error: can't parse JWT string. %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return nil, err
	}
	if !tkn.Valid {
		log.Println("error: invalid JWT token provided")
		w.WriteHeader(http.StatusUnauthorized)
		return nil, fmt.Errorf("%v", "error: invalid JWT token provided")
	}
	return claims, nil
}
