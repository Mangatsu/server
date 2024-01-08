//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package table

import (
	"github.com/go-jet/jet/v2/sqlite"
)

var User = newUserTable("", "user", "")

type userTable struct {
	sqlite.Table

	// Columns
	UUID      sqlite.ColumnString
	Username  sqlite.ColumnString
	Password  sqlite.ColumnString
	Role      sqlite.ColumnInteger
	CreatedAt sqlite.ColumnTimestamp
	UpdatedAt sqlite.ColumnTimestamp

	AllColumns     sqlite.ColumnList
	MutableColumns sqlite.ColumnList
}

type UserTable struct {
	userTable

	EXCLUDED userTable
}

// AS creates new UserTable with assigned alias
func (a UserTable) AS(alias string) *UserTable {
	return newUserTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new UserTable with assigned schema name
func (a UserTable) FromSchema(schemaName string) *UserTable {
	return newUserTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new UserTable with assigned table prefix
func (a UserTable) WithPrefix(prefix string) *UserTable {
	return newUserTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new UserTable with assigned table suffix
func (a UserTable) WithSuffix(suffix string) *UserTable {
	return newUserTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newUserTable(schemaName, tableName, alias string) *UserTable {
	return &UserTable{
		userTable: newUserTableImpl(schemaName, tableName, alias),
		EXCLUDED:  newUserTableImpl("", "excluded", ""),
	}
}

func newUserTableImpl(schemaName, tableName, alias string) userTable {
	var (
		UUIDColumn      = sqlite.StringColumn("uuid")
		UsernameColumn  = sqlite.StringColumn("username")
		PasswordColumn  = sqlite.StringColumn("password")
		RoleColumn      = sqlite.IntegerColumn("role")
		CreatedAtColumn = sqlite.TimestampColumn("created_at")
		UpdatedAtColumn = sqlite.TimestampColumn("updated_at")
		allColumns      = sqlite.ColumnList{UUIDColumn, UsernameColumn, PasswordColumn, RoleColumn, CreatedAtColumn, UpdatedAtColumn}
		mutableColumns  = sqlite.ColumnList{UsernameColumn, PasswordColumn, RoleColumn, CreatedAtColumn, UpdatedAtColumn}
	)

	return userTable{
		Table: sqlite.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		UUID:      UUIDColumn,
		Username:  UsernameColumn,
		Password:  PasswordColumn,
		Role:      RoleColumn,
		CreatedAt: CreatedAtColumn,
		UpdatedAt: UpdatedAtColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
