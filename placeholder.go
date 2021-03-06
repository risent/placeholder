package main

import (
	"code.google.com/p/freetype-go/freetype"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	// fontPath = "/usr/share/fonts/TTF/luxisr.ttf"
	// fontPath = "/usr/share/fonts/wqy-microhei/wqy-microhei.ttc"
	fontPath                 = "/usr/share/fonts/wps-office/FZYTK.TTF"
	imageQuality             = jpeg.Options{95}
	bgColor      color.Color = color.RGBA{170, 170, 170, 1}
	fontColor    color.Color = color.White
	fontSize                 = 30
	fontSpacing              = 1.5
	text         string
)

/* ************************************************************************** */
// Copy from https://code.google.com/p/gorilla/source/browse/color/hex.go

// HexModel converts any color.Color to an Hex color.
var HexModel = color.ModelFunc(hexModel)

// Hex represents an RGB color in hexadecimal format.
//
// The length must be 3 or 6 characters, preceded or not by a '#'.
type Hex string

// RGBA returns the alpha-premultiplied red, green, blue and alpha values
// for the Hex.
func (c Hex) RGBA() (uint32, uint32, uint32, uint32) {
	r, g, b := HexToRGB(c)
	return uint32(r) * 0x101, uint32(g) * 0x101, uint32(b) * 0x101, 0xffff
}

// hexModel converts a color.Color to Hex.
func hexModel(c color.Color) color.Color {
	if _, ok := c.(Hex); ok {
		return c
	}
	r, g, b, _ := c.RGBA()
	return RGBToHex(uint8(r>>8), uint8(g>>8), uint8(b>>8))
}

// RGBToHex converts an RGB triple to an Hex string.
func RGBToHex(r, g, b uint8) Hex {
	return Hex(fmt.Sprintf("#%02X%02X%02X", r, g, b))
}

// HexToRGB converts an Hex string to a RGB triple.
func HexToRGB(h Hex) (uint8, uint8, uint8) {
	if len(h) > 0 && h[0] == '#' {
		h = h[1:]
	}
	if len(h) == 3 {
		h = h[:1] + h[:1] + h[1:2] + h[1:2] + h[2:] + h[2:]
	}
	if len(h) == 6 {
		if rgb, err := strconv.ParseUint(string(h), 16, 32); err == nil {
			return uint8(rgb >> 16), uint8((rgb >> 8) & 0xFF), uint8(rgb & 0xFF)
		}
	}
	return 0, 0, 0
}

/* ************************************************************************** */

func handler(w http.ResponseWriter, r *http.Request) {
	params := strings.Split(r.URL.Path[1:], "/")
	if len(params) == 3 {
		r, g, b := HexToRGB(Hex(params[2]))
		fontColor = color.RGBA{r, g, b, 255}
	}

	if len(params) >= 2 {
		r, g, b := HexToRGB(Hex(params[1]))
		bgColor = color.RGBA{r, g, b, 255}
	}

	size := strings.Split(params[0], "x")

	var width, height int

	if len(size) == 1 {
		width, _ = strconv.Atoi(size[0])
		height = width
	} else {
		width, _ = strconv.Atoi(size[0])
		height, _ = strconv.Atoi(size[1])
	}

	texts := r.URL.Query()["text"]
	if len(texts) > 0 {
		text = texts[0]
	}

	m := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(m, m.Bounds(), &image.Uniform{bgColor}, image.ZP, draw.Src)

	fontBytes, err := ioutil.ReadFile(fontPath)
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Println(err)
		return
	}
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(font)
	c.SetFontSize(float64(fontSize))
	c.SetClip(m.Bounds())
	c.SetDst(m)
	src := image.NewUniform(fontColor)
	c.SetSrc(src)
	c.SetHinting(freetype.FullHinting)

	emSquarePix := int(c.PointToFix32(float64(fontSize)) >> 8)

	if len(text) == 0 {
		text = fmt.Sprintf("%d X %d", width, height)
	}
	textCount := utf8.RuneCountInString(text)

	x0 := (width - textCount*emSquarePix) / 2
	y0 := height / 2

	pt := freetype.Pt(x0, y0)
	c.DrawString(text, pt)
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-control", "public, max-age=259200")
	jpeg.Encode(w, m, &imageQuality)
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
