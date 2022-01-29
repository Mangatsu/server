package api

import (
	"database/sql"
	"encoding/json"
	"github.com/Mangatsu/server/internal/config"
	"github.com/Mangatsu/server/pkg/db"
	"github.com/Mangatsu/server/pkg/types/model"
	"github.com/golang-jwt/jwt/v4"
	"math"
	"net/http"
	"strconv"
	"strings"
)

type CustomClaims struct {
	jwt.RegisteredClaims
	ID      string
	Subject string
	Name    string
	Roles   *int32
}

func parseQueryParams(r *http.Request) db.Filters {
	order := db.Order(r.URL.Query().Get("order"))
	sortBy := db.SortBy(r.URL.Query().Get("sortby"))
	searchTerm := r.URL.Query().Get("search")
	category := r.URL.Query().Get("category")
	favoriteGroup := r.URL.Query().Get("favorite")
	nsfw := r.URL.Query().Get("nsfw")
	rawTags := r.URL.Query()["tag"] // namespace:name

	var tags []model.Tag
	for _, rawTag := range rawTags {
		tag := strings.Split(rawTag, ":")
		if len(tag) != 2 {
			continue
		}
		tags = append(tags, model.Tag{Namespace: tag[0], Name: tag[1]})
	}

	limit, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if err != nil {
		limit = 50
	} else {
		limit = db.Clamp(limit, 1, 100)
	}

	offset, err := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)
	if err != nil {
		offset = 0
	} else {
		offset = db.Clamp(offset, 0, math.MaxInt64)
	}

	return db.Filters{
		SearchTerm:    strings.TrimSpace(searchTerm),
		Order:         order,
		SortBy:        sortBy,
		Limit:         limit,
		Offset:        offset,
		Category:      category,
		FavoriteGroup: favoriteGroup,
		NSFW:          nsfw,
		Tags:          tags,
	}
}

func convertTagsToMap(tags []model.Tag) map[string][]string {
	tagMap := map[string][]string{}
	for _, tag := range tags {
		tagMap[tag.Namespace] = append(tagMap[tag.Namespace], tag.Name)
	}
	return tagMap
}

func convertMetadata(metadata db.CombinedMetadata) MetadataResult {
	return MetadataResult{
		ArchivePath: metadata.ArchivePath,
		Hidden:      metadata.Hidden,
		Gallery:     metadata.Gallery,
		Tags:        convertTagsToMap(metadata.Tags),
		Reference:   metadata.Reference,
		GalleryPref: metadata.GalleryPref,
		Library:     metadata.Library,
	}
}

// hasAccess handles access based on the Visibility option. Role restricts access to the specified role.
// NoRole (0) allows access to anonymous users if the Visibility is Public or Restricted (passphrase required).
func hasAccess(w http.ResponseWriter, r *http.Request, role db.Role) (bool, *string) {
	// JWT/session auth
	authorization := strings.Fields(r.Header.Get("Authorization"))
	if len(authorization) == 2 {
		access, userUUID := verifyJWT(authorization[1], role)
		if !access {
			errorHandler(w, http.StatusUnauthorized, "")
			return false, nil
		}
		return access, userUUID
	}

	// Username & password auth
	credentials := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(credentials)
	if err == nil && credentials.Username != "" && credentials.Password != "" {
		access, userUUID, _ := loginHelper(w, *credentials, role)
		if !access {
			return false, nil
		}
		return access, userUUID
	}

	// Anonymous access with passphrase. Deny if action requires any role.
	if role > 0 {
		errorHandler(w, http.StatusUnauthorized, "")
		return false, nil
	}

	// Simple global passphrase auth
	if config.CurrentVisibility() == config.Restricted {
		access := *credentials.Passphrase == config.RestrictedPassphrase()
		if !access {
			errorHandler(w, http.StatusUnauthorized, "")
			return false, nil
		}
		return access, nil
	}

	// If public, anonymous access without passphrase is allowed
	if config.CurrentVisibility() == config.Public {
		return true, nil
	}

	errorHandler(w, http.StatusUnauthorized, "")
	return false, nil
}

// loginHelper handles login
func loginHelper(w http.ResponseWriter, credentials Credentials, requiredRole db.Role) (bool, *string, *int32) {
	userUUID, role, err := db.Login(credentials.Username, credentials.Password, requiredRole)
	if err != nil || userUUID == nil {
		// If an entry with the username doesn't exist, send 401 status
		if err == sql.ErrNoRows {
			errorHandler(w, http.StatusUnauthorized, "")
			return false, nil, nil
		}
		errorHandler(w, http.StatusUnauthorized, "")
		return false, nil, nil
	}

	return true, userUUID, role
}

func newJWT(userUUID string, sessionID string, expiresIn *int64, sessionName *string, role *int32) (string, error) {
	if sessionID == "" {
		if expiresIn != nil {
			*expiresIn = db.Clamp(*expiresIn, 30, 60*60*24*30)
		}

		newSessionID, err := db.NewSession(userUUID, expiresIn, sessionName)
		if err != nil {
			return "", err
		}
		sessionID = newSessionID
	}

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, CustomClaims{
		ID:      sessionID,
		Subject: userUUID,
		Roles:   role,
	})

	token, err := claims.SignedString([]byte(config.JWTSecret()))
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
		func(token *jwt.Token) (interface{}, error) { return []byte(config.JWTSecret()), nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)

	claims, ok := token.Claims.(*CustomClaims)
	return *claims, ok, token, err
}
