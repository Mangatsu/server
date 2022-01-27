package metadata

import (
	"github.com/Luukuton/Mangatsu/pkg/library"
	"github.com/Luukuton/Mangatsu/pkg/types/model"
	"testing"
)

func TestParseTitle(t *testing.T) {
	wantMap := map[string]TitleMeta{
		"(C99) [doujin circle (some artist)] very lewd title (Magical Girls) [DL].zip": {
			Released: "C99",
			Circle:   "doujin circle",
			Artists:  "some artist",
			Title:    "very lewd title",
			Series:   "Magical Girls",
			Language: "",
		},
		"(C99) [doujin circle] very lewd title (Magical Girls) [DL].zip": {
			Released: "C99",
			Circle:   "doujin circle",
			Artists:  "",
			Title:    "very lewd title",
			Series:   "Magical Girls",
			Language: "",
		},
		"[doujin circle] very lewd title (Magical Girls) [DL].zip": {
			Released: "",
			Circle:   "doujin circle",
			Artists:  "",
			Title:    "very lewd title",
			Series:   "Magical Girls",
			Language: "",
		},
		"(C99) [doujin circle] very lewd title [DL].zip": {
			Released: "C99",
			Circle:   "doujin circle",
			Artists:  "",
			Title:    "very lewd title",
			Series:   "",
			Language: "",
		},
	}

	for title, want := range wantMap {
		got := ParseTitle(title)
		if got != want {
			t.Errorf("Parsed title (%s) didn't match the expected result", title)
		}
	}
}

func TestParseX(t *testing.T) {
	json, err := library.ReadJSON("../../testdata/x.json")
	if err != nil {
		t.Error("Reading x.json failed")
		return
	}

	exhGallery, err := unmarshalExhJSON(json)
	if err != nil {
		t.Error("Error unmarshalling JSON data:", err)
		return
	}

	archivePath := "(C99) [同人サークル (とあるアーティスト)] とてもエッチなタイトル (魔法少女) [DL版].zip"
	gotGallery, gotTags, gotExternal, err := convertExh(exhGallery, archivePath, "info.json", true)
	if err != nil {
		t.Error("Error converting Exh format:", err)
		return
	}

	if gotGallery.Title != "(C99) [doujin circle (some artist)] very lewd title (Magical Girls) [DL]" ||
		*gotGallery.TitleNative != "(C99) [同人サークル (とあるアーティスト)] とてもエッチなタイトル (魔法少女) [DL版]" ||
		*gotGallery.Category != "doujinshi" ||
		*gotGallery.Language != "Japanese" ||
		*gotGallery.Translated != false ||
		*gotGallery.ImageCount != int32(30) ||
		*gotGallery.ArchiveSize != int32(11639011) ||
		gotGallery.ArchivePath != archivePath {
		t.Error("parsed gallery didn't match the expected result")
	}

	var wantTags [3]model.Tag
	wantTags[0] = model.Tag{Namespace: "parody", Name: "Magical Girls"}
	wantTags[1] = model.Tag{Namespace: "female", Name: "swimsuit"}
	wantTags[2] = model.Tag{Namespace: "female", Name: "yuri"}

	for i, gotTag := range gotTags {
		if gotTag.Namespace != wantTags[i].Namespace || gotTag.Name != wantTags[i].Name {
			t.Error("parsed tags didn't match expected results")
		}
	}

	if *gotExternal.MetaPath != "info.json" || *gotExternal.ExhGid != int32(1) || *gotExternal.ExhToken != "abc" {
		t.Error("parsed external info didn't match the expected result")
	}
}
