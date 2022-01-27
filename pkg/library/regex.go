package library

import "regexp"

var ArchiveExtensions = regexp.MustCompile(`\.(?:zip|cbz|rar|cbr)$`)
var MetaExtensions = regexp.MustCompile(`\.(?:json|txt)$`)
var ImageExtensions = regexp.MustCompile(`\.(?:jpe?g|png|webp|bmp|gif|tiff)$`)
