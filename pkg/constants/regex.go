package constants

import "regexp"

var ArchiveExtensions = regexp.MustCompile(`\.(?:zip|cbz|rar|cbr|7z)$`)
var MetaExtensions = regexp.MustCompile(`\.(?:json|txt)$`)
var ImageExtensions = regexp.MustCompile(`\.(?:jpe?g|png|webp|avif|bmp|gif|tiff?|heif)$`)
