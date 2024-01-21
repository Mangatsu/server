package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/Mangatsu/server/internal/config"
	"github.com/Mangatsu/server/pkg/db"
	"github.com/Mangatsu/server/pkg/types/model"
	"github.com/Mangatsu/server/pkg/utils"
	"github.com/weppos/publicsuffix-go/publicsuffix"
	"math"
	"net/http"
	"strconv"
	"strings"
)

func parseQueryParams(r *http.Request) db.Filters {
	order := db.Order(r.URL.Query().Get("order"))
	sortBy := db.SortBy(r.URL.Query().Get("sortby"))
	searchTerm := r.URL.Query().Get("search")
	category := r.URL.Query().Get("category")
	series := r.URL.Query().Get("series")
	favoriteGroup := r.URL.Query().Get("favorite")
	nsfw := r.URL.Query().Get("nsfw")
	rawTags := r.URL.Query()["tag"] // namespace:name
	grouped := r.URL.Query().Get("grouped")

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
		limit = utils.Clamp(limit, 1, 100)
	}

	offset, err := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)
	if err != nil {
		offset = 0
	} else {
		offset = utils.Clamp(offset, 0, math.MaxInt64)
	}

	seed, err := strconv.ParseInt(r.URL.Query().Get("seed"), 10, 64)
	if err != nil {
		seed = 0
	}

	return db.Filters{
		SearchTerm:    strings.TrimSpace(searchTerm),
		Order:         order,
		SortBy:        sortBy,
		Limit:         limit,
		Offset:        offset,
		Category:      category,
		Series:        series,
		FavoriteGroup: favoriteGroup,
		NSFW:          nsfw,
		Tags:          tags,
		Grouped:       grouped,
		Seed:          seed,
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
	publicAccess := config.Options.Visibility == config.Public && role == db.NoRole

	token := readJWT(r)
	if token != "" {
		access, userUUID := verifyJWT(token, role)
		if access {
			return access, userUUID
		}

		if publicAccess {
			return true, nil
		}

		restrictedAccess := config.Options.Visibility == config.Restricted && role == db.NoRole
		if restrictedAccess && token == config.Credentials.Passphrase {
			return true, nil
		}

		errorHandler(w, http.StatusUnauthorized, "", r.URL.Path)
		return false, nil
	}

	// Username & password auth
	if r.Body != nil {
		credentials := &Credentials{}
		err := json.NewDecoder(r.Body).Decode(credentials)
		if err == nil && credentials.Username != "" && credentials.Password != "" {
			access, userUUID, _ := loginHelper(w, *credentials, role)
			if !access {
				errorHandler(w, http.StatusUnauthorized, "", r.URL.Path)
				return false, nil
			}
			return access, userUUID
		}
	}

	// If public, anonymous access without passphrase is allowed
	if publicAccess {
		return true, nil
	}

	errorHandler(w, http.StatusUnauthorized, "", r.URL.Path)
	return false, nil
}

// loginHelper handles login
func loginHelper(w http.ResponseWriter, credentials Credentials, requiredRole db.Role) (bool, *string, *int32) {
	err := db.MigratePassword(credentials.Username, credentials.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			errorHandler(w, http.StatusUnauthorized, "", "")
		} else {
			errorHandler(w, http.StatusInternalServerError, err.Error(), "")
		}
		return false, nil, nil
	}

	userUUID, role, err := db.Login(credentials.Username, credentials.Password, requiredRole)
	if err != nil || userUUID == nil {
		errorHandler(w, http.StatusUnauthorized, "", "")
		return false, nil, nil
	}

	return true, userUUID, role
}

// originAllowed returns true if the origin is allowed. If MTSU_STRICT_ACAO is false, it will always return true.
func originAllowed(origin string) bool {
	if !config.Options.StrictACAO {
		return true
	}

	domain, err := publicsuffix.Domain(origin)
	return err == nil && domain == config.Options.Domain
}
