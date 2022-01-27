package api

import (
	"encoding/json"
	"fmt"
	"github.com/Mangatsu/server/internal/config"
	"github.com/Mangatsu/server/pkg/db"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
)

type Credentials struct {
	Username    string  `json:"username"`
	Password    string  `json:"password"`
	Role        *string `json:"role"`
	Passphrase  *string `json:"passphrase"`
	ExpiresIn   *int64  `json:"expires_in"`
	SessionName *string `json:"session_name"`
}

func register(w http.ResponseWriter, r *http.Request) {
	credentials := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(credentials)
	if err != nil || credentials.Username == "" || credentials.Password == "" {
		errorHandler(w, http.StatusBadRequest, "")
		return
	}

	if !config.RegistrationsEnabled() || credentials.Role != nil {
		authorization := strings.Fields(r.Header.Get("Authorization"))
		if len(authorization) == 2 {
			access, _ := verifyJWT(authorization[1], db.Admin)
			if !access {
				errorHandler(w, http.StatusUnauthorized, "")
				return
			}
		}
	}

	role := int64(10)
	if credentials.Role != nil {
		role, err = strconv.ParseInt(*credentials.Role, 10, 8)
		if err != nil {
			log.Error(err)
			errorHandler(w, http.StatusBadRequest, "")
			return
		}
	}
	if err = db.Register(credentials.Username, credentials.Password, db.Clamp(role, 0, int64(db.Admin))); err != nil {
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

	access, userUUID, role := loginHelper(w, *credentials, db.Role(0))
	if !access {
		return
	}

	token, err := newJWT(*userUUID, "", credentials.ExpiresIn, credentials.SessionName, role)
	if err != nil {
		errorHandler(w, http.StatusInternalServerError, "")
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	err = json.NewEncoder(w).Encode(loginResponse{Token: token})
	if err != nil {
		errorHandler(w, http.StatusInternalServerError, "")
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	authorization := strings.Split(r.Header.Get("Authorization"), " ")
	if len(authorization) == 2 {
		claims, ok, _, err := parseJWT(authorization[1])
		if err != nil || !ok {
			errorHandler(w, http.StatusUnauthorized, "")
			return
		}

		err = db.Logout(claims.ID, claims.Subject)
		if err != nil {
			w.WriteHeader(http.StatusGone)
			fmt.Fprintf(w, `{ "code": %d, "message": "gone" }`, http.StatusGone)
			return
		}
	}
	fmt.Fprint(w, `{ "message": "successfully logged out" }`)
}

// updateUser can be used to update role, password or username of users. Role can only be changed by admins.
func updateUser(w http.ResponseWriter, r *http.Request) {
	userForm := &db.UserForm{}
	err := json.NewDecoder(r.Body).Decode(userForm)
	if err != nil {
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

	if err = db.UpdateUser(userUUID, userForm); err != nil {
		errorHandler(w, http.StatusInternalServerError, "")
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
	if err != nil {
		errorHandler(w, http.StatusInternalServerError, "")
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	err = json.NewEncoder(w).Encode(users)
	if err != nil {
		errorHandler(w, http.StatusInternalServerError, "")
		return
	}
}

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

// returnFavoriteGroups returns all user's favorite groups as JSON.
func returnFavoriteGroups(w http.ResponseWriter, r *http.Request) {
	access, userUUID := hasAccess(w, r, db.Viewer)
	if !access {
		return
	}

	favoriteGroups := db.GetFavoriteGroups(*userUUID)
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	err := json.NewEncoder(w).Encode(favoriteGroups)
	if err != nil {
		errorHandler(w, http.StatusInternalServerError, "")
		return
	}
}

func setFavorite(w http.ResponseWriter, r *http.Request) {
	access, userUUID := hasAccess(w, r, db.Viewer)
	if !access {
		return
	}

	params := mux.Vars(r)
	if params["uuid"] == "" || params["name"] == "" {
		errorHandler(w, http.StatusBadRequest, "")
		return
	}

	err := db.SetFavoriteGroup(params["name"], params["uuid"], *userUUID)
	if err != nil {
		errorHandler(w, http.StatusInternalServerError, "")
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
		errorHandler(w, http.StatusBadRequest, "")
		return
	}

	err = db.UpdateProgress(int32(progress), params["uuid"], *userUUID)

	if err != nil {
		errorHandler(w, http.StatusInternalServerError, "")
		return
	}
	fmt.Fprintf(w, `{ "message": "progress updated" }`)
}
