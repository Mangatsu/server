package db

import (
	"database/sql"
	"errors"
	"github.com/Mangatsu/server/pkg/utils"
	"math"
	"math/rand/v2"
	"time"

	"github.com/Mangatsu/server/pkg/log"
	"github.com/Mangatsu/server/pkg/types/sqlite/model"
	. "github.com/Mangatsu/server/pkg/types/sqlite/table"
	. "github.com/go-jet/jet/v2/sqlite"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CombinedMetadata struct {
	model.Gallery

	Tags []model.Tag

	Reference struct {
		ExhToken *string
		ExhGid   *int32
		Urls     *string
	} `alias:"reference.*"`

	GalleryPref *struct {
		FavoriteGroup *string
		Progress      int32
		UpdatedAt     string
	} `alias:"gallery_pref.*"`

	Library model.Library `json:"-"`
}

type MappedTags struct {
	Data  map[string][]string `json:"Data"`
	Count int
}

type Categories struct {
	Data []string `json:"Data"`
}

type Filters struct {
	Order         Order
	SortBy        SortBy
	Limit         uint64
	Offset        uint64
	SearchTerm    string
	Series        string
	Category      string
	FavoriteGroup string
	NSFW          string
	Tags          []model.Tag
	Grouped       string
	Seed          uint64
}

type SortBy string

const (
	Title       SortBy = "title"
	TitleNative        = "native"
	UpdatedAt          = "updated"
	Progress           = "progress"
)

type Order string

const (
	Desc Order = "desc"
	Asc        = "asc"
)

// NewGallery creates a new gallery
func NewGallery(archivePath string, libraryID int32, title string, series string, size int64, imageCount uint64) (string, error) {
	galleryUUID, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	now := time.Now()
	var stmt InsertStatement
	if series != "" {
		stmt = Gallery.
			INSERT(Gallery.UUID, Gallery.ArchivePath, Gallery.Title, Gallery.LibraryID, Gallery.Series, Gallery.ArchiveSize, Gallery.ImageCount, Gallery.CreatedAt, Gallery.UpdatedAt).
			VALUES(galleryUUID.String(), archivePath, title, libraryID, series, size, imageCount, now, now).
			RETURNING(Gallery.UUID)
	} else {
		stmt = Gallery.
			INSERT(Gallery.UUID, Gallery.ArchivePath, Gallery.Title, Gallery.LibraryID, Gallery.ArchiveSize, Gallery.ImageCount, Gallery.CreatedAt, Gallery.UpdatedAt).
			VALUES(galleryUUID.String(), archivePath, title, libraryID, size, imageCount, now, now).
			RETURNING(Gallery.UUID)
	}

	var galleries []model.Gallery
	err = stmt.Query(db(), &galleries)
	if err != nil {
		return "", err
	}

	if len(galleries) == 0 {
		return "", sql.ErrNoRows
	}

	return galleries[0].UUID, nil
}

// UpdateGallery updates a gallery. It also adds tags and references if any.
// If internalScan is true, the gallery is matched by its archive path, not UUID.
func UpdateGallery(gallery model.Gallery, tags []model.Tag, reference model.Reference, internalScan bool) error {
	now := time.Now()

	var tagIDs []int32
	if tags != nil {
		deleteStmt := GalleryTag.DELETE().WHERE(GalleryTag.GalleryUUID.EQ(String(gallery.UUID)))
		_, err := deleteStmt.Exec(db())
		if err != nil {
			return err
		}

		if len(tags) > 0 {
			if tagIDs, err = NewTags(tags); err != nil {
				return err
			}
		}
	}

	prevGallery, err := GetGallery(&gallery.UUID, nil, &gallery.ArchivePath)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("gallery not found")
		}
		return err
	}

	tx, err := db().Begin()
	if err != nil {
		return err
	}

	committed := false
	defer func() {
		if !committed {
			rollbackTx(tx)
		}
	}()

	var updateGalleryStmt UpdateStatement

	if internalScan {
		galleryModel, galleryColumnList := ValidateGalleryInternal(prevGallery.Gallery, now)

		updateGalleryStmt = Gallery.
			UPDATE(galleryColumnList).
			MODEL(galleryModel).
			WHERE(Gallery.ArchivePath.EQ(String(gallery.ArchivePath))).
			RETURNING(Gallery.UUID)
	} else {
		galleryModel, galleryColumnList := ValidateGallery(prevGallery.Gallery, gallery, now)

		updateGalleryStmt = Gallery.
			UPDATE(galleryColumnList).
			MODEL(galleryModel).
			WHERE(Gallery.UUID.EQ(String(gallery.UUID))).
			RETURNING(Gallery.UUID)
	}

	var galleries []model.Gallery
	if err = updateGalleryStmt.Query(tx, &galleries); err != nil {
		return err
	}
	if len(galleries) == 0 {
		return nil
	}

	// Inserts gallery tag junctions in a loop if any
	if len(tagIDs) > 0 {
		for _, tagID := range tagIDs {
			insertTagGalleryStmt := GalleryTag.
				INSERT(GalleryTag.TagID, GalleryTag.GalleryUUID).
				VALUES(tagID, galleries[0].UUID).
				ON_CONFLICT(GalleryTag.TagID, GalleryTag.GalleryUUID).
				DO_NOTHING()

			if _, err = insertTagGalleryStmt.Exec(tx); err != nil {
				return err
			}
		}
	}

	// Used for skipping title parsing if it's already done
	titleHash := utils.HashStringSHA1(gallery.Title)
	reference.MetaTitleHash = &titleHash

	// Inserts or updates reference
	referenceModel := model.Reference{}

	if internalScan {
		referenceModel = ValidateReferenceInternal(reference)
	} else {
		referenceModel = ValidateReference(reference)
	}

	referenceModel.GalleryUUID = galleries[0].UUID

	rowValues, referenceUpdateColumnList, err := ConstructExpressions(referenceModel)
	if err != nil {
		log.Z.Debug("could not construct expressions", zap.String("err", err.Error()))
		return err
	}

	insertRefStmt := Reference.
		INSERT(referenceUpdateColumnList).
		MODEL(referenceModel).
		ON_CONFLICT(Reference.GalleryUUID).
		DO_UPDATE(
			SET(referenceUpdateColumnList.
				SET(ROW(rowValues...)),
			),
		)
	if _, err = insertRefStmt.Exec(tx); err != nil {
		return err
	}

	// Commits transaction. Rollbacks on error.
	committed = true
	err = tx.Commit()
	return err
}

// NewTags creates tags from the given list.
func NewTags(tags []model.Tag) ([]int32, error) {
	var tagIDs []int32
	for _, tag := range tags {
		if tag.Namespace == "" || tag.Name == "" {
			continue
		}

		selectStmt := Tag.
			SELECT(Tag.ID, Tag.Namespace, Tag.Name).FROM(Tag).
			WHERE(Tag.Namespace.EQ(String(tag.Namespace)).AND(Tag.Name.EQ(String(tag.Name))))

		var existingTags []model.Tag
		if err := selectStmt.Query(db(), &existingTags); err != nil {
			log.Z.Debug("could not select tags, aborting", zap.String("err", err.Error()))
			return nil, err
		}

		if len(existingTags) > 0 {
			tagIDs = append(tagIDs, existingTags[0].ID)
			continue
		}

		insertStmt := Tag.INSERT(Tag.Namespace, Tag.Name).VALUES(tag.Namespace, tag.Name).RETURNING(Tag.ID)
		var insertedTags []model.Tag
		if err := insertStmt.Query(db(), &insertedTags); err != nil {
			log.Z.Debug("could not insert tags, aborting", zap.String("err", err.Error()))
			return nil, err
		}

		tagIDs = append(tagIDs, insertedTags[0].ID)
	}

	return tagIDs, nil
}

// NewGalleryPref creates initializes user preferences for a gallery.
func NewGalleryPref(galleryUUID string, userUUID string) error {
	stmt := GalleryPref.
		INSERT(GalleryPref.GalleryUUID, GalleryPref.UserUUID, GalleryPref.UpdatedAt).
		VALUES(galleryUUID, userUUID, time.Now()).
		ON_CONFLICT(GalleryPref.GalleryUUID, GalleryPref.UserUUID).DO_NOTHING()

	_, err := stmt.Exec(db())
	return err
}

// UpdateProgress sets the reading progress of a gallery for a user.
func UpdateProgress(progress int32, galleryUUID string, userUUID string) error {
	stmt := GalleryPref.
		UPDATE(GalleryPref.Progress, GalleryPref.UpdatedAt).
		SET(progress, time.Now()).
		WHERE(GalleryPref.UserUUID.EQ(String(userUUID)).AND(GalleryPref.GalleryUUID.EQ(String(galleryUUID))))

	_, err := stmt.Exec(db())
	return err
}

// SetFavoriteGroup sets a favorite group for a gallery.
func SetFavoriteGroup(favoriteGroup string, galleryUUID string, userUUID string) error {
	now := CURRENT_TIMESTAMP()
	stmt := GalleryPref.
		INSERT(GalleryPref.GalleryUUID, GalleryPref.UserUUID, GalleryPref.FavoriteGroup, GalleryPref.UpdatedAt).
		VALUES(galleryUUID, userUUID, favoriteGroup, now).
		ON_CONFLICT(GalleryPref.GalleryUUID, GalleryPref.UserUUID).
		WHERE(GalleryPref.UserUUID.EQ(String(userUUID)).AND(GalleryPref.GalleryUUID.EQ(String(galleryUUID)))).
		DO_UPDATE(SET(GalleryPref.FavoriteGroup.SET(String(favoriteGroup)), GalleryPref.UpdatedAt.SET(now)))

	_, err := stmt.Exec(db())
	return err
}

// SetThumbnail saves the filename of the thumbnail for the gallery.
func SetThumbnail(uuid string, thumbnail string) error {
	now := time.Now()
	galleryModel := model.Gallery{Thumbnail: &thumbnail, UpdatedAt: now}
	stmt := Gallery.UPDATE(Gallery.Thumbnail, Gallery.UpdatedAt).
		MODEL(galleryModel).
		WHERE(Gallery.UUID.EQ(String(uuid)))

	_, err := stmt.Exec(db())
	return err
}

func SetPageThumbnails(uuid string, pageThumbnails int32) error {
	now := time.Now()
	galleryModel := model.Gallery{PageThumbnails: &pageThumbnails, UpdatedAt: now}
	stmt := Gallery.UPDATE(Gallery.PageThumbnails, Gallery.UpdatedAt).
		MODEL(galleryModel).
		WHERE(Gallery.UUID.EQ(String(uuid)))

	_, err := stmt.Exec(db())
	return err
}

func constructGalleryFilters(filters Filters, hidden bool, userUUID *string) BoolExpression {
	// Constructing conditions
	conditions := Bool(true)

	if filters.Tags != nil {
		namespaces := make([]Expression, len(filters.Tags))
		for i, tag := range filters.Tags {
			namespaces[i] = String(tag.Namespace)
		}
		names := make([]Expression, len(filters.Tags))
		for i, tag := range filters.Tags {
			names[i] = String(tag.Name)
		}

		conditions = conditions.AND(EXISTS(SELECT(NULL).
			FROM(GalleryTag.AS("gt").INNER_JOIN(Tag.AS("t"), Tag.AS("t").ID.EQ(GalleryTag.AS("gt").TagID))).
			WHERE(
				Tag.AS("t").Namespace.IN(namespaces...).
					AND(Tag.AS("t").Name.IN(names...)).
					AND(GalleryTag.AS("gt").GalleryUUID.EQ(Gallery.UUID)),
			).GROUP_BY(GalleryTag.AS("gt").GalleryUUID).HAVING(COUNT(Tag.AS("t").Name).EQ(Int32(int32(len(filters.Tags)))))))
	}

	if filters.SearchTerm != "" {
		conditions = conditions.AND(Gallery.Title.
			LIKE(String("%" + filters.SearchTerm + "%")).
			OR(Gallery.TitleNative.LIKE(String("%" + filters.SearchTerm + "%"))).
			OR(Gallery.Series.LIKE(String("%" + filters.SearchTerm + "%"))))
	}

	if filters.Category != "" {
		conditions = conditions.AND(Gallery.Category.EQ(String(filters.Category)))
	}

	if filters.Series != "" {
		conditions = conditions.AND(Gallery.Series.EQ(String(filters.Series)))
	} else if filters.Grouped == "true" {
		conditions = conditions.AND(Gallery.Series.IS_NOT_NULL())
		conditions = conditions.AND(EXISTS(SELECT(NULL).
			FROM(Library.AS("lib")).
			WHERE(Library.AS("lib").ID.EQ(Gallery.LibraryID).
				AND(Library.AS("lib").Layout.EQ(String("structured"))),
			),
		))
	}

	if !hidden {
		conditions = conditions.AND(Gallery.Hidden.IS_NOT_TRUE())
	}

	if userUUID != nil && filters.FavoriteGroup != "" {
		conditions = conditions.AND(EXISTS(SELECT(NULL).
			FROM(GalleryPref.AS("gp")).
			WHERE(
				GalleryPref.AS("gp").GalleryUUID.EQ(Gallery.UUID).
					AND(GalleryPref.AS("gp").FavoriteGroup.EQ(String(filters.FavoriteGroup))),
			),
		))
	}

	if filters.NSFW == "false" {
		conditions = conditions.AND(Gallery.Nsfw.IS_NOT_TRUE())
	} else if filters.NSFW == "true" {
		conditions = conditions.AND(Gallery.Nsfw.IS_TRUE())
	}

	return conditions
}

// GetGalleryCount returns the number of galleries that match the given filters.
func GetGalleryCount(filters Filters, hidden bool, userUUID *string) (uint64, error) {
	stmt := Gallery.SELECT(COUNT(Gallery.UUID).AS("GalleryCount")).WHERE(constructGalleryFilters(filters, hidden, userUUID))

	var count []uint64
	err := stmt.Query(db(), &count)
	if count == nil {
		return 0, err
	}

	return count[0], err
}

// getGalleryCountHelper returns the number of galleries that match the given filters for internal use.
func getGalleryCountHelper(filters BoolExpression, uniqueSeries bool) (uint64, error) {
	var stmt SelectStatement

	if uniqueSeries {
		stmt = Gallery.SELECT(COUNT(DISTINCT(Gallery.Series)).AS("GalleryCount")).WHERE(filters)
	} else {
		stmt = Gallery.SELECT(COUNT(Gallery.UUID).AS("GalleryCount")).WHERE(filters)
	}

	var count []uint64
	err := stmt.Query(db(), &count)
	if count == nil {
		return 0, err
	}

	return count[0], err
}

// GetGalleries returns galleries based on the given filters.
func GetGalleries(filters Filters, hidden bool, userUUID *string) ([]CombinedMetadata, uint64, error) {
	doShuffle := filters.Seed != 0 && filters.Limit > 0
	conditions := constructGalleryFilters(filters, hidden, userUUID)

	var totalGalleryCount uint64
	var totalGalleryCountErr error
	if filters.Grouped == "true" {
		totalGalleryCount, totalGalleryCountErr = getGalleryCountHelper(conditions, true)
	} else {
		totalGalleryCount, totalGalleryCountErr = getGalleryCountHelper(conditions, false)
	}

	if totalGalleryCountErr != nil {
		return nil, 0, totalGalleryCountErr
	}

	filtersStmt := SELECT(Gallery.AllColumns).FROM(Gallery.Table)

	if filters.Order == Desc {
		switch filters.SortBy {
		case TitleNative:
			filtersStmt = filtersStmt.ORDER_BY(Gallery.TitleNative.DESC())
		case UpdatedAt:
			filtersStmt = filtersStmt.ORDER_BY(Gallery.UpdatedAt.DESC())
		default:
			filtersStmt = filtersStmt.ORDER_BY(Gallery.Title.DESC())
		}
	} else {
		switch filters.SortBy {
		case TitleNative:
			filtersStmt = filtersStmt.ORDER_BY(Gallery.TitleNative.ASC())
		case UpdatedAt:
			filtersStmt = filtersStmt.ORDER_BY(Gallery.UpdatedAt.ASC())
		default:
			filtersStmt = filtersStmt.ORDER_BY(Gallery.Title.ASC())

		}
	}

	var pages []uint64
	var random *rand.Rand
	if doShuffle {
		random = rand.New(rand.NewPCG(filters.Seed, 1))
		maxOffset := uint64(math.Ceil(float64(totalGalleryCount / filters.Limit)))

		pages = make([]uint64, maxOffset+1)
		for i := range pages {
			pages[i] = uint64(i)
		}

		random.Shuffle(len(pages), func(i, j int) {
			pages[i], pages[j] = pages[j], pages[i]
		})

		if filters.Offset < uint64(len(pages)) {
			filters.Offset = pages[filters.Offset]
		} else {
			filters.Offset = maxOffset + 1
		}
	}

	// Offset is multiplied by limit to get the correct offset
	filters.Offset = filters.Offset * filters.Limit

	if filters.Grouped == "true" {
		filtersStmt = filtersStmt.WHERE(conditions).LIMIT(int64(filters.Limit)).OFFSET(int64(filters.Offset)).GROUP_BY(Raw(`IFNULL(series, uuid)`))
	} else if filters.Limit == 0 {
		filtersStmt = filtersStmt.WHERE(conditions)
	} else {
		filtersStmt = filtersStmt.WHERE(conditions).LIMIT(int64(filters.Limit)).OFFSET(int64(filters.Offset))
	}

	galleryUUID := Gallery.UUID.From(filtersStmt.AsTable("galleries"))

	// Constructing statement
	joins := filtersStmt.AsTable("galleries").
		LEFT_JOIN(GalleryTag, GalleryTag.GalleryUUID.EQ(galleryUUID)).
		LEFT_JOIN(Tag, Tag.ID.EQ(GalleryTag.TagID)).
		LEFT_JOIN(Reference, Reference.GalleryUUID.EQ(galleryUUID))

	var galleriesStmt SelectStatement
	if userUUID != nil {
		galleriesStmt = SELECT(
			filtersStmt.AsTable("galleries").AllColumns(),
			Tag.Namespace,
			Tag.Name,
			Reference.ExhGid,
			Reference.ExhToken,
			Reference.Urls,
			GalleryPref.FavoriteGroup,
			GalleryPref.Progress,
			GalleryPref.UpdatedAt,
		).FROM(joins.LEFT_JOIN(
			GalleryPref, GalleryPref.GalleryUUID.EQ(galleryUUID).AND(GalleryPref.UserUUID.EQ(String(*userUUID))),
		))
	} else {
		galleriesStmt = SELECT(
			filtersStmt.AsTable("galleries").AllColumns(),
			Tag.Namespace,
			Tag.Name,
			Reference.ExhGid,
			Reference.ExhToken,
			Reference.Urls,
		).FROM(joins)
	}

	if userUUID != nil && filters.SortBy == Progress {
		if filters.Order == Desc {
			galleriesStmt = galleriesStmt.ORDER_BY(GalleryPref.Progress.DESC())
		} else {
			galleriesStmt = galleriesStmt.ORDER_BY(GalleryPref.Progress.ASC())
		}
	}

	// Shows RAW SQL-query for debugging
	//println(galleriesStmt.DebugSql())

	var galleries []CombinedMetadata
	if err := galleriesStmt.Query(db(), &galleries); err != nil {
		return nil, 0, err
	}

	if doShuffle && galleries != nil {
		random.Shuffle(len(galleries), func(i, j int) {
			galleries[i], galleries[j] = galleries[j], galleries[i]
		})
	}

	return galleries, totalGalleryCount, nil
}

// GetGallery returns a gallery based on the given UUID. If no UUID is given, a random gallery is returned.
func GetGallery(galleryUUID *string, userUUID *string, archivePath *string) (CombinedMetadata, error) {
	if (galleryUUID == nil || *galleryUUID == "") && (archivePath == nil || *archivePath == "") {
		return CombinedMetadata{}, errors.New("either galleryUUID or archivePath must be provided")
	}

	// If archivePath is provided, galleryUUID is fetched from the database.
	if (galleryUUID == nil || *galleryUUID == "") && (archivePath != nil && *archivePath != "") {
		stmt := SELECT(Gallery.UUID).FROM(Gallery).WHERE(Gallery.ArchivePath.EQ(String(*archivePath)))
		var galleries []model.Gallery
		if err := stmt.Query(db(), &galleries); err != nil {
			log.Z.Debug("could not get gallery",
				zap.Stringp("archivePath", archivePath),
				zap.String("err", err.Error()))
			return CombinedMetadata{}, err
		}

		if len(galleries) == 0 {
			log.Z.Debug("no gallery found", zap.Stringp("archivePath", archivePath))
			return CombinedMetadata{}, sql.ErrNoRows
		}

		galleryUUID = &galleries[0].UUID
	}

	// If userUUID is provided, a new user gallery entry is created.
	if userUUID != nil && galleryUUID != nil {
		if err := NewGalleryPref(*galleryUUID, *userUUID); err != nil {
			log.Z.Debug("could not add user gallery entry",
				zap.Stringp("gUUID", galleryUUID),
				zap.Stringp("uUUID", userUUID),
				zap.String("err", err.Error()))
			return CombinedMetadata{}, err
		}
	}

	// TODO: LIMIT(1) with subquery before JOINs for even better performance?
	joins := Gallery.
		LEFT_JOIN(GalleryTag, GalleryTag.GalleryUUID.EQ(Gallery.UUID)).
		LEFT_JOIN(Tag, Tag.ID.EQ(GalleryTag.TagID)).
		LEFT_JOIN(Reference, Reference.GalleryUUID.EQ(Gallery.UUID)).
		LEFT_JOIN(Library, Library.ID.EQ(Gallery.LibraryID))

	var stmt SelectStatement
	if userUUID != nil {
		joins = joins.LEFT_JOIN(
			GalleryPref, GalleryPref.GalleryUUID.EQ(Gallery.UUID).AND(GalleryPref.UserUUID.EQ(String(*userUUID))),
		)
		stmt = SELECT(
			Gallery.AllColumns,
			Tag.Namespace,
			Tag.Name,
			Reference.MetaPath,
			Reference.MetaInternal,
			Reference.ExhGid,
			Reference.ExhToken,
			Reference.Urls,
			GalleryPref.FavoriteGroup,
			GalleryPref.Progress,
			GalleryPref.UpdatedAt,
			Library.Path,
		).FROM(joins)
	} else {
		stmt = SELECT(
			Gallery.AllColumns,
			Tag.Namespace,
			Tag.Name,
			Reference.MetaPath,
			Reference.MetaInternal,
			Reference.ExhGid,
			Reference.ExhToken,
			Reference.Urls,
			Library.Path,
		).FROM(joins)
	}

	if galleryUUID != nil {
		stmt = stmt.WHERE(Gallery.UUID.EQ(String(*galleryUUID)))
	} else {
		stmt = stmt.WHERE(Gallery.UUID.IN(Raw("(SELECT gallery.uuid FROM gallery ORDER BY RANDOM() LIMIT 1)")))
	}

	var galleries []CombinedMetadata
	if err := stmt.Query(db(), &galleries); err != nil {
		log.Z.Debug("could not get gallery",
			zap.Stringp("gUUID", galleryUUID),
			zap.String("err", err.Error()))
		return CombinedMetadata{}, err
	}

	if len(galleries) == 0 {
		log.Z.Debug("no gallery found", zap.Stringp("uuid", galleryUUID))
		return CombinedMetadata{}, sql.ErrNoRows
	}

	// When requesting a random gallery and if needed, a new user gallery entry is created after the SELECT.
	if userUUID != nil && galleryUUID == nil {
		if err := NewGalleryPref(galleries[0].UUID, *userUUID); err != nil {
			log.Z.Debug("could not add user gallery entry when random gallery was requested",
				zap.Stringp("gUUID", galleryUUID),
				zap.Stringp("uUUID", userUUID),
				zap.String("err", err.Error()))
			return CombinedMetadata{}, err
		}
	}

	return galleries[0], nil
}

func GetReference(galleryUUID string) (model.Reference, error) {
	stmt := SELECT(Reference.AllColumns).
		FROM(Gallery.LEFT_JOIN(Reference, Reference.GalleryUUID.EQ(String(galleryUUID))))

	var references []model.Reference
	err := stmt.Query(db(), &references)
	if err != nil || len(references) == 0 {
		return model.Reference{}, err
	}

	return references[0], nil
}

// GetTags returns all tags.
func GetTags(galleryUUID string, mapped bool) (MappedTags, []model.Tag, error) {
	var stmt SelectStatement
	if galleryUUID != "" {
		stmt = SELECT(Tag.Namespace, Tag.Name).FROM(Gallery.
			LEFT_JOIN(GalleryTag, GalleryTag.GalleryUUID.EQ(String(galleryUUID))).
			LEFT_JOIN(Tag, Tag.ID.EQ(GalleryTag.TagID)),
		)
	} else {
		stmt = SELECT(Tag.Namespace, Tag.Name).FROM(Tag)
	}

	var tags []model.Tag
	err := stmt.Query(db(), &tags)
	if err != nil {
		return MappedTags{Data: map[string][]string{}, Count: 0}, nil, err
	}

	if !mapped {
		return MappedTags{Data: map[string][]string{}, Count: 0}, tags, err
	}

	tagMap := map[string][]string{}
	for _, tag := range tags {
		tagMap[tag.Namespace] = append(tagMap[tag.Namespace], tag.Name)
	}

	return MappedTags{Data: tagMap, Count: len(tags)}, nil, err
}

// GetCategories returns all public categories.
func GetCategories() ([]string, error) {
	stmt := SELECT(Gallery.Category).DISTINCT().FROM(Gallery.Table)
	var categories []string

	err := stmt.Query(db(), &categories)
	return categories, err
}

// GetSeries returns all series.
func GetSeries() ([]string, error) {
	stmt := SELECT(Gallery.Series).DISTINCT().FROM(Gallery.Table)
	var series []string

	err := stmt.Query(db(), &series)
	return series, err
}

// TitleHashMatch returns true if the title hash of the gallery matches the stored hash.
func TitleHashMatch(galleryUUID string) bool {
	stmt := SELECT(Gallery.UUID.AS("UUID"), Gallery.Title.AS("Title"), Reference.MetaTitleHash.AS("MetaTitleHash")).
		FROM(Gallery.Table.
			LEFT_JOIN(Reference, Reference.GalleryUUID.EQ(String(galleryUUID))),
		).
		WHERE(Gallery.UUID.EQ(String(galleryUUID))).
		LIMIT(1)

	var galleries []struct {
		UUID          string
		Title         string
		MetaTitleHash *string
	}
	err := stmt.Query(db(), &galleries)
	if err != nil {
		log.Z.Debug("failed to query for title hash",
			zap.String("uuid", galleryUUID),
			zap.String("err", err.Error()))
		return false
	}

	if galleries == nil || galleries[0].MetaTitleHash == nil {
		return false
	}

	titleHash := utils.HashStringSHA1(galleries[0].Title)

	return titleHash == *galleries[0].MetaTitleHash
}

// NeedsUpdate returns true if the gallery needs to be updated. Currently only the timestamp is checked.
func NeedsUpdate(archivePath string, updatedAt time.Time) (bool, string) {
	// TODO: Possibly also check for ArchiveSize and ArchiveHash?
	stmt := SELECT(Gallery.AllColumns).FROM(Gallery.Table).WHERE(Gallery.ArchivePath.EQ(String(archivePath)))

	var galleries []CombinedMetadata
	err := stmt.Query(db(), &galleries)
	if err != nil {
		log.Z.Error(err.Error())
		return false, ""
	}

	if galleries[0].UpdatedAt.Before(updatedAt) {
		return true, galleries[0].UUID
	}

	return false, ""
}

// ArchivePathFound returns true if the given archive path is already in the database.
func ArchivePathFound(archivePath string) []model.Gallery {
	stmt := SELECT(Gallery.UUID, Gallery.ArchivePath).
		FROM(Gallery.Table).
		WHERE(Gallery.ArchivePath.EQ(String(archivePath)))

	var galleries []model.Gallery
	err := stmt.Query(db(), &galleries)
	if err != nil {
		log.Z.Debug("failed to query for archive path",
			zap.String("archivePath", archivePath),
			zap.String("err", err.Error()))
		return nil
	}

	// ArchivePath is unique, so there should only be one result
	return galleries
}

// MetaPathFound returns true if a gallery with the given meta path exists.
func MetaPathFound(metaPath string, libraryPath string) bool {
	stmt := SELECT(Reference.MetaPath, Library.Path).
		FROM(Gallery.
			INNER_JOIN(Reference, Reference.GalleryUUID.EQ(Gallery.UUID)).
			INNER_JOIN(Library, Library.ID.EQ(Gallery.LibraryID))).
		WHERE(Reference.MetaPath.EQ(String(metaPath)).AND(Library.Path.EQ(String(libraryPath))))

	var galleries []struct{ Reference []struct{ model.Reference } }
	err := stmt.Query(db(), &galleries)
	if err != nil {
		log.Z.Debug("failed to query for meta path",
			zap.String("metaPath", metaPath),
			zap.String("err", err.Error()))
		return false
	}

	return len(galleries) > 0
}

// DeleteGallery removes a gallery from the database (never actually deletes the archive file).
func DeleteGallery(galleryUUID string) bool {
	stmt := Gallery.DELETE().WHERE(Gallery.UUID.EQ(String(galleryUUID)))
	res, err := stmt.Exec(db())
	if err != nil {
		log.Z.Debug("failed to delete a gallery",
			zap.String("uuid", galleryUUID),
			zap.String("err", err.Error()))
		return false
	}

	rowsAffected, _ := res.RowsAffected()
	return rowsAffected > int64(0)
}

func CountAllImages(skipWithPageThumbnails bool) (int, int, error) {
	stmt := SELECT(COUNT(Gallery.UUID).AS("CoverCount"), SUM(Gallery.ImageCount).AS("ImageCount")).
		FROM(Gallery.Table)

	if skipWithPageThumbnails {
		stmt = stmt.WHERE(Gallery.PageThumbnails.GT(Int32(0)))
	}

	var counts struct {
		CoverCount int
		ImageCount int
	}

	err := stmt.Query(db(), &counts)
	if err != nil {
		return 0, 0, err
	}

	return counts.CoverCount, counts.ImageCount, nil
}
