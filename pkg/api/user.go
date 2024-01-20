package api

import (
	"encoding/json"
	"fmt"
	"github.com/Mangatsu/server/internal/config"
	"github.com/Mangatsu/server/pkg/db"
	"github.com/Mangatsu/server/pkg/log"
	"github.com/Mangatsu/server/pkg/types/model"
	"github.com/Mangatsu/server/pkg/utils"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type Credentials struct {
	Username    string  `json:"username"`
	Password    string  `json:"password"`
	Passphrase  string  `json:"passphrase"`
	Role        *string `json:"role"`
	ExpiresIn   *int64  `json:"expires_in"`
	SessionName *string `json:"session_name"`
}

type LoginResponse struct {
	UUID      *string
	Role      *int32
	ExpiresIn *int64
}

const yearInSeconds = 365 * 24 * 60 * 60

func register(w http.ResponseWriter, r *http.Request) {
	credentials := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(credentials)
	if err != nil || credentials.Username == "" || credentials.Password == "" {
		errorHandler(w, http.StatusBadRequest, "", r.URL.Path)
		return
	}

	if !config.Options.Registrations || credentials.Role != nil {
		token := readJWT(r)
		if token == "" {
			errorHandler(w, http.StatusBadRequest, "", r.URL.Path)
			return
		}
		if access, _ := verifyJWT(token, db.Admin); !access {
			errorHandler(w, http.StatusUnauthorized, "", r.URL.Path)
			return
		}
	}

	role := int64(10)
	if credentials.Role != nil {
		role, err = strconv.ParseInt(*credentials.Role, 10, 8)
		if err != nil {
			errorHandler(w, http.StatusBadRequest, "failed to parse role value when registering: "+err.Error(), r.URL.Path)
			return
		}
	}
	if err = db.Register(credentials.Username, credentials.Password, utils.Clamp(role, 0, int64(db.Admin))); err != nil {
		w.WriteHeader(http.StatusConflict)
		fmt.Fprintf(w, `{ "Username already in use" }`)
		return
	}
	fmt.Fprintf(w, `{ "message": "Successfully registered" }`)
}

func login(w http.ResponseWriter, r *http.Request) {
	credentials := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(credentials)
	if err != nil {
		errorHandler(w, http.StatusBadRequest, "", r.URL.Path)
		return
	}

	if credentials.Username != "" && credentials.Password != "" {
		access, userUUID, role := loginHelper(w, *credentials, db.Role(0))
		if !access {
			return
		}

		token, err := newJWT(*userUUID, credentials.ExpiresIn, credentials.SessionName, role)
		if err != nil {
			errorHandler(w, http.StatusInternalServerError, err.Error(), r.URL.Path)
			return
		}

		jwtCookie := http.Cookie{
			Name:     "mtsu.jwt",
			Value:    "Bearer " + token,
			Domain:   config.Options.Domain,
			Path:     "/",
			MaxAge:   yearInSeconds,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteNoneMode,
		}

		http.SetCookie(w, &jwtCookie)

		resultToJSON(w, LoginResponse{
			UUID:      userUUID,
			Role:      role,
			ExpiresIn: credentials.ExpiresIn,
		}, r.URL.Path)
		return
	} else if credentials.Passphrase == config.Credentials.Passphrase {
		passphraseCookie := http.Cookie{
			Name:     "mtsu.jwt",
			Value:    "Passphrase " + credentials.Passphrase,
			Domain:   config.Options.Domain,
			Path:     "/",
			MaxAge:   yearInSeconds,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteNoneMode,
		}
		http.SetCookie(w, &passphraseCookie)

		expiresIn := int64(yearInSeconds)
		resultToJSON(w, LoginResponse{
			UUID:      nil,
			Role:      nil,
			ExpiresIn: &expiresIn,
		}, r.URL.Path)
		return
	}

	errorHandler(w, http.StatusBadRequest, "", r.URL.Path)
}

func logout(w http.ResponseWriter, r *http.Request) {
	token := readJWT(r)
	if token != "" {
		claims, ok, _, _ := parseJWT(token)

		if ok && claims.ID != "" && claims.Subject != "" && claims.Roles != nil {
			if err := db.Logout(claims.ID, claims.Subject); err != nil {
				log.Z.Debug("failed to logout", zap.String("err", err.Error()))
			}
		}
	}

	cookie := http.Cookie{
		Name:     "mtsu.jwt",
		Value:    "",
		Domain:   config.Options.Domain,
		Path:     "/",
		MaxAge:   0,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	}
	http.SetCookie(w, &cookie)

	fmt.Fprint(w, `{ "message": "successfully logged out" }`)
}

// updateUser can be used to update role, password or username of users. Role can only be changed by admins.
func updateUser(w http.ResponseWriter, r *http.Request) {
	userForm := &db.UserForm{}
	if err := json.NewDecoder(r.Body).Decode(userForm); err != nil {
		errorHandler(w, http.StatusBadRequest, err.Error(), r.URL.Path)
		return
	}

	params := mux.Vars(r)
	userUUID := params["uuid"]
	if userUUID == "" {
		errorHandler(w, http.StatusBadRequest, "", r.URL.Path)
		return
	}

	if userForm.Role != nil {
		access, _ := hasAccess(w, r, db.Admin)
		if !access {
			return
		}
	} else {
		access, loggedInUUID := hasAccess(w, r, db.Viewer)
		if *loggedInUUID != userUUID {
			errorHandler(w, http.StatusUnauthorized, "", r.URL.Path)
			return
		}
		if !access {
			return
		}
	}

	if err := db.UpdateUser(userUUID, userForm); err != nil {
		errorHandler(w, http.StatusInternalServerError, err.Error(), r.URL.Path)
		return
	}

	fmt.Fprint(w, `{ "message": "successfully updated user" }`)
}

// returnUsers returns all users in the database. Only for admins. Never returns the hashed password.
func returnUsers(w http.ResponseWriter, r *http.Request) {
	access, _ := hasAccess(w, r, db.Admin)
	if !access {
		return
	}

	users, err := db.GetUsers()
	if handleResult(w, users, err, true, r.URL.Path) {
		return
	}

	resultToJSON(w, struct {
		Data  []model.User
		Count int
	}{
		Data:  users,
		Count: len(users),
	}, r.URL.Path)
}

// deleteUser deletes a user from the database. Only for admins.
func deleteUser(w http.ResponseWriter, r *http.Request) {
	access, _ := hasAccess(w, r, db.Admin)
	if !access {
		return
	}

	params := mux.Vars(r)
	userUUID := params["uuid"]
	err := db.DeleteUser(userUUID)
	if err != nil {
		errorHandler(w, http.StatusBadRequest, err.Error(), r.URL.Path)
		return
	}
}

// returnSessions returns all sessions of the user.
func returnSessions(w http.ResponseWriter, r *http.Request) {
	access, userUUID := hasAccess(w, r, db.Viewer)
	if !access {
		return
	}
	if userUUID == nil {
		errorHandler(w, http.StatusBadRequest, "", r.URL.Path)
		return
	}

	sessions, err := db.GetSessions(*userUUID)
	if handleResult(w, sessions, err, true, r.URL.Path) {
		return
	}

	claims, _, _, err := parseJWT(readJWT(r))
	if err != nil {
		errorHandler(w, http.StatusInternalServerError, err.Error(), r.URL.Path)
		return
	}

	resultToJSON(w, struct {
		Data           []model.Session
		CurrentSession string
		Count          int
	}{
		Data:           sessions,
		CurrentSession: claims.ID,
		Count:          len(sessions),
	}, r.URL.Path)
}

// deleteSession deletes user's session from the database.
func deleteSession(w http.ResponseWriter, r *http.Request) {
	access, userUUID := hasAccess(w, r, db.Viewer)
	if !access {
		return
	}
	if userUUID == nil {
		errorHandler(w, http.StatusBadRequest, "", r.URL.Path)
		return
	}

	credentials := &struct{ SessionID string }{}
	if err := json.NewDecoder(r.Body).Decode(credentials); err != nil {
		errorHandler(w, http.StatusBadRequest, "", r.URL.Path)
		return
	}

	if err := db.DeleteSession(credentials.SessionID, *userUUID); err != nil {
		errorHandler(w, http.StatusBadRequest, err.Error(), r.URL.Path)
		return
	}
}

// returnFavoriteGroups returns all user's favorite groups as JSON.
func returnFavoriteGroups(w http.ResponseWriter, r *http.Request) {
	access, userUUID := hasAccess(w, r, db.Viewer)
	if !access {
		return
	}

	favoriteGroups, err := db.GetFavoriteGroups(*userUUID)
	if handleResult(w, favoriteGroups, err, true, r.URL.Path) {
		return
	}

	resultToJSON(w, GenericStringResult{
		Data:  favoriteGroups,
		Count: len(favoriteGroups),
	}, r.URL.Path)
}

// setFavorite sets a personal favorite group for a gallery.
func setFavorite(w http.ResponseWriter, r *http.Request) {
	access, userUUID := hasAccess(w, r, db.Viewer)
	if !access {
		return
	}

	params := mux.Vars(r)
	if params["uuid"] == "" {
		errorHandler(w, http.StatusBadRequest, "uuid is required", r.URL.Path)
		return
	}

	if err := db.SetFavoriteGroup(params["name"], params["uuid"], *userUUID); err != nil {
		errorHandler(w, http.StatusInternalServerError, err.Error(), r.URL.Path)
		return
	}
	fmt.Fprintf(w, `{ "message": "favorite group updated" }`)
}

func updateProgress(w http.ResponseWriter, r *http.Request) {
	access, userUUID := hasAccess(w, r, db.Viewer)
	if !access {
		return
	}

	params := mux.Vars(r)
	progress, err := strconv.ParseInt(params["progress"], 10, 32)
	if err != nil || params["uuid"] == "" {
		errorHandler(w, http.StatusBadRequest, err.Error(), r.URL.Path)
		return
	}

	if err = db.UpdateProgress(int32(progress), params["uuid"], *userUUID); err != nil {
		errorHandler(w, http.StatusInternalServerError, err.Error(), r.URL.Path)
		return
	}
	fmt.Fprintf(w, `{ "message": "progress updated" }`)
}
