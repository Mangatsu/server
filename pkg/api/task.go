package api

import (
	"fmt"
	"github.com/Mangatsu/server/pkg/cache"
	"github.com/Mangatsu/server/pkg/db"
	"github.com/Mangatsu/server/pkg/library"
	"github.com/Mangatsu/server/pkg/metadata"
	"net/http"
)

func scanLibraries(w http.ResponseWriter, r *http.Request) {
	access, _ := hasAccess(w, r, db.Admin)
	if !access {
		return
	}

	// fullScan := r.URL.Query().Get("full")
	go library.ScanArchives()

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	fmt.Fprintf(w, `{ "message": "Started scanning for new archives." }`)
}

func returnProcessingStatus(w http.ResponseWriter, r *http.Request) {
	access, _ := hasAccess(w, r, db.Admin)
	if !access {
		return
	}

	status := cache.ProcessingStatusCache
	resultToJSON(w, status, r.URL.Path)
}

func generateThumbnails(w http.ResponseWriter, r *http.Request) {
	access, _ := hasAccess(w, r, db.Admin)
	if !access {
		return
	}

	pages := r.URL.Query().Get("pages")
	force := r.URL.Query().Get("force")
	go library.GenerateThumbnails(pages == "true", force == "true")

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	fmt.Fprintf(w, `{ "message": "Started generateting thumbnails. Prioritizing covers." }`)
}

func findMetadata(w http.ResponseWriter, r *http.Request) {
	access, _ := hasAccess(w, r, db.Admin)
	if !access {
		return
	}

	title := r.URL.Query().Get("title")
	x := r.URL.Query().Get("x")
	ehdl := r.URL.Query().Get("ehdl")
	hath := r.URL.Query().Get("hath")

	metaTypes := make(map[metadata.MetaType]bool)
	metaTypes[metadata.XMeta] = x == "true"
	metaTypes[metadata.EHDLMeta] = ehdl == "true"
	metaTypes[metadata.HathMeta] = hath == "true"
	go metadata.ParseMetadata(metaTypes)

	if title == "true" {
		go metadata.ParseTitles(true, false)
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	if metaTypes[metadata.XMeta] || metaTypes[metadata.EHDLMeta] || metaTypes[metadata.HathMeta] || title == "true" {
		fmt.Fprintf(w, `{ "message": "started parsing given sources" }`)
		return
	}

	errorHandler(w, http.StatusBadRequest, "no sources specified", r.URL.Path)
}
