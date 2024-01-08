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

var Reference = newReferenceTable("", "reference", "")

type referenceTable struct {
	sqlite.Table

	// Columns
	GalleryUUID  sqlite.ColumnString
	MetaInternal sqlite.ColumnBool
	MetaPath     sqlite.ColumnString
	MetaMatch    sqlite.ColumnInteger
	Urls         sqlite.ColumnString
	ExhGid       sqlite.ColumnInteger
	ExhToken     sqlite.ColumnString
	AnilistID    sqlite.ColumnInteger

	AllColumns     sqlite.ColumnList
	MutableColumns sqlite.ColumnList
}

type ReferenceTable struct {
	referenceTable

	EXCLUDED referenceTable
}

// AS creates new ReferenceTable with assigned alias
func (a ReferenceTable) AS(alias string) *ReferenceTable {
	return newReferenceTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new ReferenceTable with assigned schema name
func (a ReferenceTable) FromSchema(schemaName string) *ReferenceTable {
	return newReferenceTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new ReferenceTable with assigned table prefix
func (a ReferenceTable) WithPrefix(prefix string) *ReferenceTable {
	return newReferenceTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new ReferenceTable with assigned table suffix
func (a ReferenceTable) WithSuffix(suffix string) *ReferenceTable {
	return newReferenceTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newReferenceTable(schemaName, tableName, alias string) *ReferenceTable {
	return &ReferenceTable{
		referenceTable: newReferenceTableImpl(schemaName, tableName, alias),
		EXCLUDED:       newReferenceTableImpl("", "excluded", ""),
	}
}

func newReferenceTableImpl(schemaName, tableName, alias string) referenceTable {
	var (
		GalleryUUIDColumn  = sqlite.StringColumn("gallery_uuid")
		MetaInternalColumn = sqlite.BoolColumn("meta_internal")
		MetaPathColumn     = sqlite.StringColumn("meta_path")
		MetaMatchColumn    = sqlite.IntegerColumn("meta_match")
		UrlsColumn         = sqlite.StringColumn("urls")
		ExhGidColumn       = sqlite.IntegerColumn("exh_gid")
		ExhTokenColumn     = sqlite.StringColumn("exh_token")
		AnilistIDColumn    = sqlite.IntegerColumn("anilist_id")
		allColumns         = sqlite.ColumnList{GalleryUUIDColumn, MetaInternalColumn, MetaPathColumn, MetaMatchColumn, UrlsColumn, ExhGidColumn, ExhTokenColumn, AnilistIDColumn}
		mutableColumns     = sqlite.ColumnList{MetaInternalColumn, MetaPathColumn, MetaMatchColumn, UrlsColumn, ExhGidColumn, ExhTokenColumn, AnilistIDColumn}
	)

	return referenceTable{
		Table: sqlite.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		GalleryUUID:  GalleryUUIDColumn,
		MetaInternal: MetaInternalColumn,
		MetaPath:     MetaPathColumn,
		MetaMatch:    MetaMatchColumn,
		Urls:         UrlsColumn,
		ExhGid:       ExhGidColumn,
		ExhToken:     ExhTokenColumn,
		AnilistID:    AnilistIDColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
