package utils

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Mangatsu/server/internal/config"
	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"image"
	"image/draw"
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

const DEFAULT_ASPECT_RATIO = 1.5
const DEFAULT_THUMB_COVER_WIDTH = 512
const DEFAULT_THUMB_PAGE_WIDTH = 256
const DEFAULT_COVER_WIDTH_SHIFT = 0.025

// EncodeImage encodes the given image to the format specified in the config.
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

// TransformImage resizes the given image to the specified width and returns it.
// If cover is true, the image will be tried to be cropped to the specified width as some covers include the back cover as well.
func TransformImage(srcImage *image.NRGBA, width int, cropLandscape bool, ltr bool) *image.NRGBA {
	var dstImage *image.NRGBA

	// For covers. If the image is wider than it is tall, crops the image to the specified width
	if cropLandscape {

		shift := int(float64(srcImage.Bounds().Dx()) * DEFAULT_COVER_WIDTH_SHIFT)
		width := int(float64(srcImage.Bounds().Dy()) / DEFAULT_ASPECT_RATIO)
		if srcImage.Bounds().Dx() < width {
			width = srcImage.Bounds().Dx()
		}

		var startX int
		var endX int

		if ltr {
			endX = srcImage.Bounds().Dx() / 2
			startX = endX - width

			startX -= shift
			endX -= shift
		} else {
			startX = srcImage.Bounds().Dx() / 2
			endX = startX + width

			startX += shift
			endX += shift
		}

		// Ensures the cropping window is within the image bounds
		if startX < 0 {
			startX = 0
		}
		if endX > srcImage.Bounds().Dx() {
			endX = srcImage.Bounds().Dx()
		}

		// Crops the image
		dstImage = imaging.Crop(srcImage, image.Rect(startX, 0, endX, srcImage.Bounds().Dy()))
	}

	if dstImage == nil {
		dstImage = imaging.Resize(srcImage, width, 0, imaging.Lanczos)
	} else {
		dstImage = imaging.Resize(dstImage, width, 0, imaging.Lanczos)
	}

	if dstImage.Bounds().Dy() > (int(float64(dstImage.Bounds().Dx()) * DEFAULT_ASPECT_RATIO)) {
		// If height is more than the aspect ratio allows, crops the top of the image to match the aspect ratio
		dstImage = imaging.Crop(
			dstImage,
			image.Rect(0, 0, dstImage.Bounds().Dx(), int(float64(dstImage.Bounds().Dx())*DEFAULT_ASPECT_RATIO)),
		)
	}

	return dstImage
}

func ConvertImageToNRGBA(srcImage image.Image) (*image.NRGBA, error) {
	var nrgba *image.NRGBA

	switch img := srcImage.(type) {
	case *image.NRGBA:
		nrgba = img
	case
		*image.NRGBA64,
		*image.YCbCr,
		*image.NYCbCrA,
		*image.RGBA,
		*image.RGBA64,
		*image.Gray,
		*image.Gray16,
		*image.CMYK,
		*image.Paletted:

		b := img.Bounds()
		nrgba = image.NewNRGBA(b)
		draw.Draw(nrgba, b, img, b.Min, draw.Src)
	default:
		return nil, errors.New(fmt.Sprintf("unsupported image type: %T", img))
	}

	return nrgba, nil
}
