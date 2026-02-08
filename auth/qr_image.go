package auth

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strings"
)

const pngDataURLPrefix = "data:image/png;base64,"

var errInvalidPNGDataURL = errors.New("invalid png data url")

// normalizePNGDataURL ensures the input is a decodable PNG data URL and that the
// decoded image is square (Slack QR decode logic requires square bounds).
//
// If the image is not square, it pads the shorter side with white background and
// centers the original.
func normalizePNGDataURL(s string) (string, error) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(strings.ToLower(s), pngDataURLPrefix) {
		return "", errInvalidPNGDataURL
	}

	raw, err := base64.StdEncoding.DecodeString(s[len(pngDataURLPrefix):])
	if err != nil {
		return "", err
	}
	img, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		return "", err
	}

	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w == h {
		// Return trimmed original (keeps size under slackauth max input limit).
		return s, nil
	}

	n := w
	if h > n {
		n = h
	}

	dst := image.NewRGBA(image.Rect(0, 0, n, n))
	draw.Draw(dst, dst.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)

	offX := (n - w) / 2
	offY := (n - h) / 2
	draw.Draw(dst, image.Rect(offX, offY, offX+w, offY+h), img, b.Min, draw.Over)

	var buf bytes.Buffer
	if err := png.Encode(&buf, dst); err != nil {
		return "", err
	}
	return pngDataURLPrefix + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

