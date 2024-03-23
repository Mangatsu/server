package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Mangatsu/server/internal/config"
	"github.com/Mangatsu/server/pkg/db"
	"github.com/Mangatsu/server/pkg/log"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"net/http"
	"reflect"
	"time"
)

// handleResult handles the result and returns if it was successful or not.
// InternalServerError will be set if any error found. NotFound is set if the result is nil or empty.
func handleResult(w http.ResponseWriter, result interface{}, err error, many bool, endpoint string) bool {
	resultType := reflect.TypeOf(result)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, qrm.ErrNoRows) {
			errorHandler(w, http.StatusNotFound, "", endpoint)
			return true
		} else {
			errorHandler(w, http.StatusInternalServerError, err.Error(), endpoint)
			return true
		}
	}
	if !many {
		if result == nil {
			errorHandler(w, http.StatusNotFound, "", endpoint)
			return true
		}
		if resultType.Kind() == reflect.Slice {
			if resultType.Kind() == reflect.Ptr {
				result = reflect.ValueOf(result).Elem().Interface()
			}
			list := reflect.ValueOf(result)
			if list.Len() == 0 {
				errorHandler(w, http.StatusNotFound, "", endpoint)
				return true
			}
		}
	}
	return false
}

func resultToJSON(w http.ResponseWriter, result interface{}, endpoint string) {
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		errorHandler(w, http.StatusInternalServerError, err.Error(), endpoint)
	}
}

func returnInfo(w http.ResponseWriter, r *http.Request) {
	resultToJSON(w, struct {
		APIVersion    int
		ServerVersion string
		Visibility    config.Visibility
		Registrations bool
	}{
		APIVersion:    1,
		ServerVersion: "0.8.0",
		Visibility:    config.Options.Visibility,
		Registrations: config.Options.Registrations,
	}, r.URL.Path)
}

// Returns statistics as JSON.
func returnStatistics(w http.ResponseWriter, r *http.Request) {
	if access, _ := hasAccess(w, r, db.NoRole); !access {
		return
	}

	fmt.Fprintf(w, `{ "message": "statistics not implemented" }`)
}

// Returns the root path as JSON.
func returnRoot(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, `{ "message": "Mangatsu API available at /api" }`)
}

// Handles errors. Argument msg is only used with 400 and 500.
func errorHandler(w http.ResponseWriter, status int, msg string, endpoint string) {
	switch status {
	case http.StatusNotFound:
		w.WriteHeader(status)
		fmt.Fprintf(w, `{ "code": %d, "message": "not found" }`, status)
	case http.StatusBadRequest:
		w.WriteHeader(status)
		fmt.Fprintf(w, `{ "code": %d, "message": "%s" }`, status, msg)
	case http.StatusForbidden:
		w.WriteHeader(status)
		fmt.Fprintf(w, `{ "code": %d, "message": "forbidden" }`, status)
	case http.StatusUnauthorized:
		w.WriteHeader(status)
		fmt.Fprintf(w, `{ "code": %d, "message": "unauthorized" }`, status)
	case http.StatusGone:
		w.WriteHeader(status)
		fmt.Fprintf(w, `{ "code": %d, "message": "gone" }`, status)
	case http.StatusConflict:
		w.WriteHeader(status)
		fmt.Fprintf(w, `{ "code": %d, "message": "%s" }`, status, msg)
	default:
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{ "status": %d, "message": "internal server error" }`, http.StatusInternalServerError)
		log.Z.Error(msg, zap.Int("status", status))
		return
	}

	log.Z.Debug(
		"api request",
		zap.String("endpoint", endpoint),
		zap.Int("status", status),
		zap.String("msg", msg),
	)
}

// Handles HTTP(S) requests.
func handleRequests() {
	baseURL := "/api/v1"
	uuidRegex := "[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}"
	r := mux.NewRouter().StrictSlash(true)

	r.HandleFunc("/", returnRoot).Methods("GET")
	r.HandleFunc("/api", returnInfo).Methods("GET")
	r.HandleFunc(baseURL+"/statistics", returnStatistics).Methods("GET")

	r.HandleFunc(baseURL+"/register", register).Methods("POST")
	r.HandleFunc(baseURL+"/login", login).Methods("POST")
	r.HandleFunc(baseURL+"/logout", logout).Methods("POST")

	r.HandleFunc(baseURL+"/users", returnUsers).Methods("GET")
	r.HandleFunc(baseURL+"/users/{uuid:"+uuidRegex+"}", updateUser).Methods("PUT")
	r.HandleFunc(baseURL+"/users/{uuid:"+uuidRegex+"}", deleteUser).Methods("DELETE")
	r.HandleFunc(baseURL+"/users/me/favorites", returnFavoriteGroups).Methods("GET")
	r.HandleFunc(baseURL+"/users/me/sessions", returnSessions).Methods("GET")
	r.HandleFunc(baseURL+"/users/me/sessions", deleteSession).Methods("DELETE")

	r.HandleFunc(baseURL+"/status", returnProcessingStatus).Methods("GET")
	r.HandleFunc(baseURL+"/scan", scanLibraries).Methods("GET")
	r.HandleFunc(baseURL+"/thumbnails", generateThumbnails).Methods("GET")
	r.HandleFunc(baseURL+"/meta", findMetadata).Methods("GET")

	r.HandleFunc(baseURL+"/categories", returnCategories).Methods("GET")
	r.HandleFunc(baseURL+"/series", returnSeries).Methods("GET")
	r.HandleFunc(baseURL+"/tags", returnTags).Methods("GET")

	r.HandleFunc(baseURL+"/galleries", returnGalleries).Methods("GET")
	r.HandleFunc(baseURL+"/galleries/count", returnGalleryCount).Methods("GET")
	r.HandleFunc(baseURL+"/galleries/random", returnRandomGallery).Methods("GET")
	r.HandleFunc(baseURL+"/galleries/{uuid:"+uuidRegex+"}", updateGallery).Methods("PUT")
	r.HandleFunc(baseURL+"/galleries/{uuid:"+uuidRegex+"}", returnGallery).Methods("GET")
	r.HandleFunc(baseURL+"/galleries/{uuid:"+uuidRegex+"}/progress/{progress:[0-9]+}", updateProgress).Methods("PATCH")
	r.HandleFunc(baseURL+"/galleries/{uuid:"+uuidRegex+"}/favorite/{name}", setFavorite).Methods("PATCH")
	r.HandleFunc(baseURL+"/galleries/{uuid:"+uuidRegex+"}/favorite", setFavorite).Methods("PATCH")

	if config.Options.Cache.WebServer {
		r.PathPrefix("/cache/").Handler(http.StripPrefix("/cache/", http.FileServer(http.Dir(config.BuildCachePath()))))
	}

	// General 404
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorHandler(w, http.StatusNotFound, "", r.RequestURI)
	})

	handler := cors.New(cors.Options{
		AllowOriginFunc: func(origin string) bool { return originAllowed(origin) },
		AllowedMethods: []string{
			http.MethodOptions, http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPut, http.MethodPatch,
		},
		AllowedHeaders: []string{
			"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization",
			"Access-Control-Allow-Headers", "Origin", "X-Requested-With", "Access-Control-Request-Method",
			"Access-Control-Request-Headers",
		},
		AllowCredentials:    true,
		AllowPrivateNetwork: true,
	}).Handler(r)

	fullAddress := config.Options.Hostname + ":" + config.Options.Port
	srv := &http.Server{
		Handler:      handler,
		Addr:         fullAddress,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Z.Info("starting API on: " + fullAddress)
	log.Z.Warn(srv.ListenAndServe().Error())
}

// LaunchAPI starts handling API requests.
func LaunchAPI() {
	handleRequests()
}
