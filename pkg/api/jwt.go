package api

import (
	"net/http"
	"strings"

	"github.com/Mangatsu/server/internal/config"
	"github.com/Mangatsu/server/pkg/db"
	"github.com/Mangatsu/server/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	jwt.RegisteredClaims
	ID      string
	Subject string
	Name    string
	Roles   *int32
}

// readJWT parses the JWT from an HTTP request's Cookie or Authorization header.
func readJWT(r *http.Request) string {
	jwtCookie, err := r.Cookie("mtsu.jwt")   // Mostly for web browsers
	jwtAuth := r.Header.Get("Authorization") // Others such as mobile apps

	token := ""
	if err == nil {
		token = jwtCookie.Value
	} else {
		token = jwtAuth
	}

	splitToken := strings.Fields(token)
	if len(splitToken) == 2 {
		return splitToken[1]
	}

	return ""
}

func newJWT(userUUID string, expiresIn *int64, sessionName *string, role *int32) (string, error) {
	if expiresIn != nil {
		*expiresIn = utils.Clamp(*expiresIn, 30, 60*60*24*365)
	}

	sessionID, err := db.NewSession(userUUID, expiresIn, sessionName)
	if err != nil {
		return "", err
	}

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, CustomClaims{
		ID:      sessionID,
		Subject: userUUID,
		Roles:   role,
	})

	token, err := claims.SignedString([]byte(config.Credentials.JWTSecret))
	if err != nil {
		return "", err
	}

	return token, err
}

func verifyJWT(tokenString string, role db.Role) (bool, *string) {
	claims, ok, token, err := parseJWT(tokenString)

	claimedRole := 0
	if claims.Roles != nil {
		claimedRole = int(*claims.Roles)
	}

	if err == nil && ok && token.Valid && db.VerifySession(claims.ID, claims.Subject) && claimedRole >= int(role) {
		return true, &claims.Subject
	}

	return false, nil
}

func parseJWT(tokenString string) (CustomClaims, bool, *jwt.Token, error) {
	token, err := jwt.ParseWithClaims(
		tokenString, &CustomClaims{},
		func(token *jwt.Token) (interface{}, error) { return []byte(config.Credentials.JWTSecret), nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
	if err != nil {
		return CustomClaims{}, false, nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	return *claims, ok, token, err
}
