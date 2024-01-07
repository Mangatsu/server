package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Mangatsu/server/internal/config"
	"github.com/Mangatsu/server/pkg/db"
	"github.com/Mangatsu/server/pkg/log"
	"github.com/Mangatsu/server/pkg/types/model"
	"github.com/Mangatsu/server/pkg/utils"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
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
		errorHandler(w, http.StatusBadRequest, "")
		return
	}

	if !config.Options.Registrations || credentials.Role != nil {
		token := readJWT(r)
		if token == "" {
			errorHandler(w, http.StatusBadRequest, "")
			return
		}
		if access, _ := verifyJWT(token, db.Admin); !access {
			errorHandler(w, http.StatusUnauthorized, "")
			return
		}
	}

	role := int64(10)
	if credentials.Role != nil {
		role, err = strconv.ParseInt(*credentials.Role, 10, 8)
		if err != nil {
			log.Z.Debug("failed to parse role value when registering", zap.String("err", err.Error()))
			errorHandler(w, http.StatusBadRequest, "")
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
		errorHandler(w, http.StatusBadRequest, "")
		return
	}

	if credentials.Username != "" && credentials.Password != "" {
		access, userUUID, role := loginHelper(w, *credentials, db.Role(0))
		if !access {
			return
		}

		token, err := newJWT(*userUUID, credentials.ExpiresIn, credentials.SessionName, role)
		if err != nil {
			errorHandler(w, http.StatusInternalServerError, err.Error())
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
		})
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

		resultToJSON(w, LoginResponse{
			UUID:      nil,
			Role:      nil,
			ExpiresIn: nil,
		})
		return
	}

	errorHandler(w, http.StatusBadRequest, "")
}

func logout(w http.ResponseWriter, r *http.Request) {
	token := readJWT(r)
	if token == "" {
		errorHandler(w, http.StatusBadRequest, "")
		return
	}

	claims, ok, _, err := parseJWT(token)
	if err != nil || !ok {
		errorHandler(w, http.StatusUnauthorized, "")
		return
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

	if err = db.Logout(claims.ID, claims.Subject); err != nil {
		w.WriteHeader(http.StatusGone)
		fmt.Fprintf(w, `{ "code": %d, "message": "gone" }`, http.StatusGone)
		return
	}

	fmt.Fprint(w, `{ "message": "successfully logged out" }`)
}

// updateUser can be used to update role, password or username of users. Role can only be changed by admins.
func updateUser(w http.ResponseWriter, r *http.Request) {
	userForm := &db.UserForm{}
	if err := json.NewDecoder(r.Body).Decode(userForm); err != nil {
		errorHandler(w, http.StatusBadRequest, "")
		return
	}

	params := mux.Vars(r)
	userUUID := params["uuid"]
	if userUUID == "" {
		errorHandler(w, http.StatusBadRequest, "")
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
			errorHandler(w, http.StatusUnauthorized, "")
			return
		}
		if !access {
			return
		}
	}

	if err := db.UpdateUser(userUUID, userForm); err != nil {
		errorHandler(w, http.StatusInternalServerError, err.Error())
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
	if handleResult(w, users, err, true) {
		return
	}

	resultToJSON(w, struct {
		Data  []model.User
		Count int
	}{
		Data:  users,
		Count: len(users),
	})
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
		errorHandler(w, http.StatusBadRequest, "")
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
		errorHandler(w, http.StatusBadRequest, "")
		return
	}

	sessions, err := db.GetSessions(*userUUID)
	if handleResult(w, sessions, err, true) {
		return
	}

	claims, _, _, err := parseJWT(readJWT(r))
	if err != nil {
		errorHandler(w, http.StatusInternalServerError, "")
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
	})
}

// deleteSession deletes user's session from the database.
func deleteSession(w http.ResponseWriter, r *http.Request) {
	access, userUUID := hasAccess(w, r, db.Viewer)
	if !access {
		return
	}
	if userUUID == nil {
		errorHandler(w, http.StatusBadRequest, "")
		return
	}

	credentials := &struct{ SessionID string }{}
	if err := json.NewDecoder(r.Body).Decode(credentials); err != nil {
		errorHandler(w, http.StatusBadRequest, "")
		return
	}

	if err := db.DeleteSession(credentials.SessionID, *userUUID); err != nil {
		errorHandler(w, http.StatusBadRequest, "")
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
	if handleResult(w, favoriteGroups, err, true) {
		return
	}

	resultToJSON(w, GenericStringResult{
		Data:  favoriteGroups,
		Count: len(favoriteGroups),
	})
}

// setFavorite sets a personal favorite group for a gallery.
func setFavorite(w http.ResponseWriter, r *http.Request) {
	access, userUUID := hasAccess(w, r, db.Viewer)
	if !access {
		return
	}

	params := mux.Vars(r)
	if params["uuid"] == "" {
		errorHandler(w, http.StatusBadRequest, "")
		return
	}

	if err := db.SetFavoriteGroup(params["name"], params["uuid"], *userUUID); err != nil {
		errorHandler(w, http.StatusInternalServerError, err.Error())
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
		errorHandler(w, http.StatusBadRequest, err.Error())
		return
	}

	if err = db.UpdateProgress(int32(progress), params["uuid"], *userUUID); err != nil {
		errorHandler(w, http.StatusInternalServerError, err.Error())
		return
	}
	fmt.Fprintf(w, `{ "message": "progress updated" }`)
}
