package metadata

import (
	"github.com/Mangatsu/server/pkg/utils"
	"os"
	"testing"
)

func TestParseTitle(t *testing.T) {
	wantMap := map[string]TitleMeta{
		"(C99) [doujin circle (some artist)] very lewd title (Magical Girls) [DL].zip": {
			Released: "C99",
			Circle:   "doujin circle",
			Artists:  []string{"some artist"},
			Title:    "very lewd title",
			Series:   "Magical Girls",
			Language: "",
		},
		"(C99) [doujin circle] very lewd title (Magical Girls) [DL].zip": {
			Released: "C99",
			Circle:   "doujin circle",
			Artists:  []string{""},
			Title:    "very lewd title",
			Series:   "Magical Girls",
			Language: "",
		},
		"[doujin circle] very lewd title (Magical Girls) [DL].zip": {
			Released: "",
			Circle:   "doujin circle",
			Artists:  []string{""},
			Title:    "very lewd title",
			Series:   "Magical Girls",
			Language: "",
		},
		"(C99) [doujin circle] very lewd title [DL].zip": {
			Released: "C99",
			Circle:   "doujin circle",
			Artists:  []string{""},
			Title:    "very lewd title",
			Series:   "",
			Language: "",
		},
	}

	for title, want := range wantMap {
		got := ParseTitle(title)
		if got.Released != want.Released &&
			got.Circle != want.Circle &&
			got.Title != want.Title &&
			got.Series != want.Series &&
			got.Language != want.Language {
			t.Errorf("Parsed title (%s) didn't match the expected result", title)
		}
		for i, artist := range got.Artists {
			if artist != want.Artists[i] {
				t.Errorf("Parsed title's (%s) artists didn't match the expected result", title)
			}
		}
	}
}

func TestParseX(t *testing.T) {
	json, err := utils.ReadJSON("../../testdata/x.json")
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
	gotGallery, gotTags, gotReference := convertExh(exhGallery, archivePath, "info.json", true)

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

	wantTags := map[string]string{}
	wantTags["Magical Girls"] = "parody"
	wantTags["swimsuit"] = "female"
	wantTags["yuri"] = "female"

	for _, gotTag := range gotTags {

		if wantTags[gotTag.Name] != gotTag.Namespace {
			t.Error("parsed tags didn't match expected results: ", wantTags[gotTag.Namespace], " - ", gotTag.Name)
		}
	}

	if *gotReference.MetaPath != "info.json" || *gotReference.ExhGid != int32(1) || *gotReference.ExhToken != "abc" {
		t.Error("parsed reference info didn't match the expected result")
	}
}

func TestParseHath(t *testing.T) {
	filepath := "../../testdata/hath.txt"

	buf, err := os.ReadFile(filepath)
	if err != nil {
		t.Error("Error reading hath.txt:", err)
		return
	}

	gotGallery, gotTags, _, err := ParseHath(filepath, buf, false)
	if err != nil {
		t.Error("Error parsing galleryinfo.txt:", err)
		return
	}

	if *gotGallery.TitleNative != "(C88) [hサークル] とてもエッチなタイトル (魔法少女)" {
		t.Error("parsed gallery didn't match the expected result")
	}

	wantTags := map[string]string{}
	wantTags["mahou shoujo"] = "parody"
	wantTags["hcircle"] = "group"
	wantTags["group"] = "female"
	wantTags["thigh high boots"] = "female"
	wantTags["artbook"] = "other"

	for _, gotTag := range gotTags {
		if wantTags[gotTag.Name] != gotTag.Namespace {
			t.Error("parsed tags didn't match expected results: ", wantTags[gotTag.Namespace], " - ", gotTag.Name)
		}
	}
}

func TestParseEHDL(t *testing.T) {
	filepath := "../../testdata/ehdl.txt"

	buf, err := os.ReadFile(filepath)
	if err != nil {
		t.Error("Error reading ehdl.txt:", err)
		return
	}

	gotGallery, gotTags, _, err := ParseEHDL(filepath, buf, false)
	if err != nil {
		t.Error("Error parsing ehdl.txt:", err)
		return
	}

	println("gotGallery.ArchiveSize:", *gotGallery.ArchiveSize)
	if gotGallery.Title != "[CRAZY CIRCLE (Hana)] Oppai Oppai Oppai" ||
		*gotGallery.TitleNative != "[CRAZY CIRCLE (はな)] おっぱいおっぱいおっぱい" ||
		*gotGallery.Category != "doujinshi" ||
		*gotGallery.Language != "Japanese" ||
		*gotGallery.ImageCount != int32(12) ||
		*gotGallery.ArchiveSize != int32(69690002) {
		t.Error("parsed gallery didn't match the expected result")
	}

	wantTags := map[string]string{}
	wantTags["crazy circle"] = "group"
	wantTags["artist"] = "hana"
	wantTags["group"] = "female"
	wantTags["fft threesome"] = "female"
	wantTags["stockings"] = "female"

	for _, gotTag := range gotTags {
		if wantTags[gotTag.Name] != gotTag.Namespace {
			t.Error("parsed tags didn't match expected results: ", wantTags[gotTag.Namespace], " - ", gotTag.Name)
		}
	}
}
