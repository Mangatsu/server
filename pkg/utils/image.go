package utils

import (
	"bytes"
	"errors"
	"github.com/Mangatsu/server/internal/config"
	"github.com/chai2010/webp"
	"image"
)

// TODO: test how long it takes to generate webp thumbnails compared to jpg + size differences
//newImage, _ := os.Create("../cache/thumbnails/" + galleryUUID + "/" + name)
//defer func(newImage *os.File) {
//	err := newImage.Close()
//	if err != nil {
//		log.Error("Closing thumbnail: ", err)
//	}
//}(newImage)
//err = png.Encode(newImage, dstImage)
//if err != nil {
//	log.Error("Writing thumbnail: ", err)
//	return
//}

func EncodeImage(dstImage *image.NRGBA) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	var err error

	switch config.Options.GalleryOptions.ThumbnailFormat {
	case config.WebP:
		err = webp.Encode(&buf, dstImage, &webp.Options{Lossless: false, Quality: 75})
	case config.AVIF:
		return nil, errors.New("avif not supported yet")
	default:
		return nil, errors.New("unknown image format")
	}

	return &buf, err
}
