package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"github.com/Mangatsu/server/pkg/types/model"
	. "github.com/Mangatsu/server/pkg/types/table"
	"github.com/Mangatsu/server/pkg/utility"
	. "github.com/go-jet/jet/v2/sqlite"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"io"
	"strconv"
	"time"
)

type UserForm struct {
	Password *string `json:"password"`
	Username *string `json:"username"`
	Role     *string `json:"role"`
}

type FavoriteGroups struct {
	Data []string `json:"Data"`
}

type Role int8

const (
	Admin  Role = 100
	Member      = 20
	Viewer      = 10
	NoRole      = 0
)

// GetUser returns a user from the database.
func GetUser(name string) ([]model.User, error) {
	stmt := SELECT(
		User.AllColumns,
	).FROM(
		User.Table,
	).WHERE(
		User.Username.EQ(String(name)),
	)

	var user []model.User
	err := stmt.Query(db(), &user)
	return user, err
}

// GetUsers returns users from the database.
func GetUsers() ([]model.User, error) {
	stmt := SELECT(
		User.UUID,
		User.Username,
		User.Role,
		User.CreatedAt,
		User.UpdatedAt,
	).FROM(
		User.Table,
	)

	var users []model.User
	err := stmt.Query(db(), &users)
	return users, err
}

// GetFavoriteGroups returns user's favorite groups.
func GetFavoriteGroups(userUUID string) ([]string, error) {
	stmt := SELECT(GalleryPref.FavoriteGroup).DISTINCT().
		FROM(GalleryPref.Table).
		WHERE(GalleryPref.UserUUID.EQ(String(userUUID)).
			AND(GalleryPref.FavoriteGroup.IS_NOT_NULL()).
			AND(GalleryPref.FavoriteGroup.NOT_EQ(String(""))))

	var favoriteGroups []string
	err := stmt.Query(db(), &favoriteGroups)
	return favoriteGroups, err
}

// Register registers a new user.
func Register(username string, password string, role int64) error {
	now := time.Now()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	userUUID, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	insertUser := User.
		INSERT(User.UUID, User.Username, User.Password, User.Role, User.CreatedAt, User.UpdatedAt).
		VALUES(userUUID.String(), username, hashedPassword, role, now, now)

	_, err = insertUser.Exec(db())
	return err
}

// Login logs the user in and returns the UUID of the user.
func Login(username string, password string, role Role) (*string, *int32, error) {
	result, err := GetUser(username)
	if err != nil {
		return nil, nil, err
	}

	// No user found
	if len(result) == 0 {
		return nil, nil, sql.ErrNoRows
	}

	if Role(result[0].Role) < role {
		return nil, nil, nil
	}

	if err = bcrypt.CompareHashAndPassword([]byte(result[0].Password), []byte(password)); err != nil {
		return nil, nil, err
	}

	return &result[0].UUID, &result[0].Role, err
}

// Logout logs out a user by removing a session.
func Logout(sessionUUID string, userUUID string) error {
	return DeleteSession(sessionUUID, userUUID)
}

// NewSession creates a new session for a user.
func NewSession(userUUID string, expiresIn *int64, sessionName *string) (string, error) {
	x := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, x); err != nil {
		return "", err
	}
	sessionID := base64.URLEncoding.EncodeToString(x)

	if expiresIn != nil {
		expiresAt := time.Now().Add(time.Duration(*expiresIn) * time.Second).Unix()
		expiresIn = &expiresAt
	}

	stmt := Session.
		INSERT(Session.ID, Session.UserUUID, Session.Name, Session.ExpiresAt).
		VALUES(sessionID, userUUID, sessionName, expiresIn)

	_, err := stmt.Exec(db())
	return sessionID, err
}

// VerifySession verifies a session by checking if it exists based on the session ID and user UUID.
func VerifySession(id string, userUUID string) bool {
	stmt := SELECT(Session.ID).
		FROM(Session.Table).
		WHERE(Session.ID.EQ(String(id)).AND(Session.UserUUID.EQ(String(userUUID)))).
		LIMIT(1)

	var sessions []model.Session
	err := stmt.Query(db(), &sessions)
	if err != nil {
		log.Debug("Error verifying session: ", err)
		return false
	}

	return len(sessions) > 0
}

// UpdateUser can be used to update role, password or username of users.
func UpdateUser(userUUID string, userForm *UserForm) error {
	now := time.Now()

	tx, err := db().Begin()
	if err != nil {
		return err
	}
	defer rollbackTx(tx)

	if userForm.Role != nil {
		role, err := strconv.ParseInt(*userForm.Role, 10, 8)
		if err != nil {
			return err
		}
		role = utility.Clamp(role, NoRole, int64(Admin))

		updateUserStmt := User.
			UPDATE(User.Role, User.UpdatedAt).
			SET(role, now).
			WHERE(User.UUID.EQ(String(userUUID)))
		if _, err = updateUserStmt.Exec(tx); err != nil {
			return err
		}

		// If role is changed, delete all sessions of the user.
		deleteSessionsStmt := Session.DELETE().WHERE(Session.UserUUID.EQ(String(userUUID)))
		if _, err = deleteSessionsStmt.Exec(tx); err != nil {
			return err
		}
	}

	if userForm.Username != nil && *userForm.Username != "" {
		updateUserStmt := User.
			UPDATE(User.Username, User.UpdatedAt).
			SET(userForm.Username, now).
			WHERE(User.UUID.EQ(String(userUUID)))
		if _, err = updateUserStmt.Exec(tx); err != nil {
			return err
		}
	}

	if userForm.Password != nil && *userForm.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*userForm.Password), 12)
		updateUserStmt := User.
			UPDATE(User.Password, User.UpdatedAt).
			SET(hashedPassword, now).
			WHERE(User.UUID.EQ(String(userUUID)))
		if _, err = updateUserStmt.Exec(tx); err != nil {
			return err
		}
	}

	// Commit transaction. Rollback on error.
	err = tx.Commit()
	return err
}

// DeleteUser removes user. Admins cannot be deleted, they have to demoted first.
func DeleteUser(userUUID string) error {
	stmt := User.DELETE().WHERE(User.UUID.EQ(String(userUUID)).AND(User.Role.NOT_EQ(Int8(int8(Admin)))))
	_, err := stmt.Exec(db())
	return err
}

// GetSessions returns all sessions of a user.
func GetSessions(userUUID string) ([]model.Session, error) {
	stmt := Session.SELECT(Session.AllColumns).WHERE(Session.UserUUID.EQ(String(userUUID)))
	var sessions []model.Session
	err := stmt.Query(db(), &sessions)
	return sessions, err
}

// DeleteSession removes a session based on the session ID and user UUID.
func DeleteSession(id string, userUUID string) error {
	stmt := Session.DELETE().WHERE(Session.ID.EQ(String(id)).AND(Session.UserUUID.EQ(String(userUUID))))
	_, err := stmt.Exec(db())
	return err
}

// PruneExpiredSessions removes all expired sessions.
func PruneExpiredSessions() {
	stmt := Session.DELETE().WHERE(BoolExp(Raw("unixepoch() > session.expires_at")))
	if _, err := stmt.Exec(db()); err != nil {
		log.Error("Error pruning expired sessions: ", err)
	}
}
