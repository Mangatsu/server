package cache

type processingError struct {
	UUIDOrPath string
	Error      string
}

type scanResult struct {
	Running          bool
	FoundGalleries   []string
	SkippedGalleries []string
	Errors           []processingError
}

type thumbnailResult struct {
	Running         bool
	GeneratedCovers int
	GeneratedPages  int
	Errors          []processingError
}

type metadataResult struct {
	Running bool
	Errors  []processingError
}

type ProcessingStatus struct {
	Scan       scanResult
	Thumbnails thumbnailResult
	Metadata   metadataResult
}

var ProcessingStatusCache *ProcessingStatus

func InitProcessingStatusCache() {
	ProcessingStatusCache = &ProcessingStatus{
		Scan: scanResult{
			Running:          false,
			FoundGalleries:   make([]string, 0),
			SkippedGalleries: make([]string, 0),
			Errors:           make([]processingError, 0),
		},
		Thumbnails: thumbnailResult{
			Running:         false,
			GeneratedCovers: 0,
			GeneratedPages:  0,
			Errors:          make([]processingError, 0),
		},
		Metadata: metadataResult{
			Running: false,
			Errors:  make([]processingError, 0),
		},
	}
}

func (s *ProcessingStatus) SetScanRunning(running bool) {
	s.Scan.Running = running
}

func (s *ProcessingStatus) AddScanFoundGallery(galleryUUID string) {
	s.Scan.FoundGalleries = append(s.Scan.FoundGalleries, galleryUUID)
}

func (s *ProcessingStatus) AddScanSkippedGallery(galleryUUID string) {
	s.Scan.SkippedGalleries = append(s.Scan.SkippedGalleries, galleryUUID)
}

func (s *ProcessingStatus) AddScanError(uuidOrPath string, err string) {
	s.Scan.Errors = append(s.Scan.Errors, processingError{
		UUIDOrPath: uuidOrPath,
		Error:      err,
	})
}

func (s *ProcessingStatus) SetThumbnailsRunning(running bool) {
	s.Thumbnails.Running = running
}

func (s *ProcessingStatus) AddThumbnailGeneratedCover() {
	s.Thumbnails.GeneratedCovers++
}

func (s *ProcessingStatus) AddThumbnailGeneratedPage() {
	s.Thumbnails.GeneratedPages++
}

func (s *ProcessingStatus) AddThumbnailError(uuidOrPath string, err string) {
	s.Thumbnails.Errors = append(s.Thumbnails.Errors, processingError{
		UUIDOrPath: uuidOrPath,
		Error:      err,
	})
}

func (s *ProcessingStatus) SetMetadataRunning(running bool) {
	s.Metadata.Running = running

}

func (s *ProcessingStatus) AddMetadataError(uuidOrPath string, err string) {
	s.Metadata.Errors = append(s.Metadata.Errors, processingError{
		UUIDOrPath: uuidOrPath,
		Error:      err,
	})
}

func (s *ProcessingStatus) Reset() {
	s.Scan = scanResult{
		Running:          false,
		FoundGalleries:   make([]string, 0),
		SkippedGalleries: make([]string, 0),
		Errors:           make([]processingError, 0),
	}
	s.Thumbnails = thumbnailResult{
		Running:         false,
		GeneratedCovers: 0,
		GeneratedPages:  0,
		Errors:          make([]processingError, 0),
	}
	s.Metadata = metadataResult{
		Running: false,
		Errors:  make([]processingError, 0),
	}
}
