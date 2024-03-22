package utils

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/Mangatsu/server/pkg/log"
	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"time"
)

// ReadJSON returns the given JSON file as bytes.
func ReadJSON(jsonFile string) ([]byte, error) {
	jsonFileBytes, err := os.ReadFile(jsonFile)

	if err != nil {
		log.Z.Debug("failed to read JSON file", zap.String("err", err.Error()))
		return nil, err
	}

	return jsonFileBytes, nil
}

// Clamp clamps the given value to the given range.
func Clamp(value, min, max int64) int64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// ClampU clamps the given unsigned value to the given range.
func ClampU(value, min, max uint64) uint64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// PeriodicTask loops the given function in separate thread between the given interval.
func PeriodicTask(d time.Duration, f func()) {
	go func() {
		for {
			f()
			time.Sleep(d)
		}
	}()
}

// PathExists checks if the given path exists.
func PathExists(pathTo string) bool {
	_, err := os.Stat(pathTo)
	if os.IsNotExist(err) {
		return false
	}

	return err == nil
}

func FileSize(filePath string) (int64, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		log.Z.Error("could not get file size", zap.String("path", filePath), zap.String("err", err.Error()))
		return 0, err
	}

	return stat.Size(), nil
}

func DirSize(dirPath string) (int64, error) {
	var size int64
	err := filepath.Walk(dirPath, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

// Similarity calculates the similarity between two strings.
func Similarity(a string, b string) float64 {
	sd := metrics.NewSorensenDice()
	sd.CaseSensitive = false
	sd.NgramSize = 4
	similarity := strutil.Similarity(a, b, sd)

	return similarity
}

func HashStringSHA1(s string) string {
	h := sha1.New()
	h.Write([]byte(s))

	return hex.EncodeToString(h.Sum(nil))
}
