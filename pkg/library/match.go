package library

import (
	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
)

func Similarity(a string, b string) float64 {
	sd := metrics.NewSorensenDice()
	sd.CaseSensitive = false
	sd.NgramSize = 4
	similarity := strutil.Similarity(a, b, sd)

	return similarity
}
