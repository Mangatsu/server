package api

import (
	"fmt"
	"github.com/Mangatsu/server/pkg/db"
	"github.com/Mangatsu/server/pkg/library"
	"github.com/Mangatsu/server/pkg/metadata"
	"net/http"
	"strings"
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

	x := r.URL.Query().Get("x")
	title := r.URL.Query().Get("title")
	var sources []string

	if x == "true" {
		go metadata.ParseX()
		sources = append(sources, "X JSON")
	}
	if title == "true" {
		go metadata.ParseTitles(true, false)
		sources = append(sources, "titles")
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	if len(sources) > 0 {
		response := strings.Join(sources, ",")
		fmt.Fprintf(w, `{ "message": "Started parsing given sources: `+response+`" }`)
		return
	}

	errorHandler(w, http.StatusBadRequest, "No sources specified.")
}
