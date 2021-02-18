package controllers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"golang.org/x/image/draw"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-zoo/bone"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/psi-incontrol/go-sprocket/page"
	"github.com/psi-incontrol/go-sprocket/sprocket"
	"golang.org/x/image/font"
)

// PageController ...
type PageController struct {
	fontFile string
}

// RequestMapping ...
func (z *PageController) RequestMapping(router *bone.Mux) {
	router.GetFunc("/", http.HandlerFunc(z.TextToImageHandler))
	router.PostFunc("/", http.HandlerFunc(z.TextToImageHandler))
}

func (z *PageController) loadFont() (*truetype.Font, error) {
	z.fontFile = "static/fonts/UbuntuMono-R.ttf"
	fontBytes, err := ioutil.ReadFile(z.fontFile)
	if err != nil {
		return nil, err
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (z *PageController) generateImage(textContent string, fgColorHex string, bgColorHex string, fontSize float64) ([]byte, error) {

	fgColor := color.RGBA{0xff, 0xff, 0xff, 0xff}
	if len(fgColorHex) == 7 {
		_, err := fmt.Sscanf(fgColorHex, "#%02x%02x%02x", &fgColor.R, &fgColor.G, &fgColor.B)
		if err != nil {
			log.Println(err)
			fgColor = color.RGBA{0x2e, 0x34, 0x36, 0xff}
		}
	}

	bgColor := color.RGBA{0x30, 0x0a, 0x24, 0xff}
	if len(bgColorHex) == 7 {
		_, err := fmt.Sscanf(bgColorHex, "#%02x%02x%02x", &bgColor.R, &bgColor.G, &bgColor.B)
		if err != nil {
			log.Println(err)
			bgColor = color.RGBA{0x30, 0x0a, 0x24, 0xff}
		}
	}

	loadedFont, err := z.loadFont()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	code := strings.Replace(textContent, "\t", "    ", -1) // convert tabs into spaces
	text := strings.Split(code, "\n") // split newlines into arrays

	fg := image.NewUniform(fgColor)
	bg := image.NewUniform(bgColor)
	rgba := image.NewRGBA(image.Rect(0, 0, 1200, 630))
	draw.Draw(rgba, rgba.Bounds(), bg, image.Pt(0, 0), draw.Src)
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(loadedFont)
	c.SetFontSize(fontSize)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(fg)
	c.SetHinting(font.HintingNone)

	textXOffset := 50
	textYOffset := 10+int(c.PointToFixed(fontSize)>>6) // Note shift/truncate 6 bits first

	pt := freetype.Pt(textXOffset, textYOffset)
	for _, s := range text {
		_, err = c.DrawString(strings.Replace(s, "\r", "", -1), pt)
		if err != nil {
			return nil, err
		}
		pt.Y += c.PointToFixed(fontSize * 1.5)
	}

	b := new(bytes.Buffer)
	if err := png.Encode(b, rgba); err != nil {
		log.Println("unable to encode image.")
		return nil, err
	}
	return b.Bytes(), nil
}

// TextToImageHandler ...
func (z *PageController) TextToImageHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		{
			page := page.New()
			page.Title = "Generate Text to Image in Go"
			renderPage(w, r, page, "views/base.html", "views/text-to-image-builder.html")
		}
	case http.MethodPost:
		{
			textContent := strings.Trim(r.FormValue("textContent"), " ")

			fontSize, err := strconv.ParseFloat(strings.Trim(r.FormValue("fontSize"), " "), 64)
			if err != nil {
				fontSize = 32
			}

			fgColorHex := strings.Trim(r.FormValue("fgColor"), " ")
			fgColorHex = strings.ToLower(fgColorHex)

			bgColorHex := strings.Trim(r.FormValue("bgColor"), " ")
			bgColorHex = strings.ToLower(bgColorHex)

			b, err := z.generateImage(textContent, fgColorHex, bgColorHex, fontSize)
			if err != nil {
				log.Println(err)
				sprocket.RespondStatusCodeWithJSON(w, 500, nil)
				return
			}

			str := base64.StdEncoding.EncodeToString(b)
			ppage := page.New()
			ppage.Title = "Generate Text to Image in Go"

			data := make(map[string]interface{})
			data["generatedImage"] = str

			ppage.SetData(data)
			renderPage(w, r, ppage, "views/base.html", "views/text-to-image-result.html")
		}
	default:
		{
		}
	}
}
