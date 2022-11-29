package main

import (
	"bytes"
	"embed"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/johnpili/go-text-to-image/models"
	"github.com/johnpili/go-text-to-image/page"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/image/font"
	"gopkg.in/yaml.v2"
	"html/template"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	configuration models.Config
	fontFile      string

	//go:embed static/*
	static embed.FS

	//go:embed views/*
	views embed.FS
)

func main() {
	pid := os.Getpid()
	err := os.WriteFile("application.pid", []byte(strconv.Itoa(pid)), 0666)
	if err != nil {
		log.Fatal(err)
	}

	var configLocation string
	flag.StringVar(&configLocation, "config", "config.yml", "Set the location of configuration file")
	flag.Parse()

	loadConfiguration(configLocation, &configuration)

	f := fs.FS(static)
	v, _ := fs.Sub(f, "static")

	router := httprouter.New()
	router.ServeFiles("/static/*filepath", http.FS(v))
	router.HandlerFunc("GET", "/", indexHandler)
	router.HandlerFunc("POST", "/", indexHandler)

	port := strconv.Itoa(configuration.HTTP.Port)
	httpServer := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
	}

	if configuration.HTTP.IsTLS {
		log.Printf("Server running at https://localhost/tools:%s/\n", port)
		log.Fatal(httpServer.ListenAndServeTLS(configuration.HTTP.ServerCert, configuration.HTTP.ServerKey))
		return
	}
	log.Printf("Server running at http://localhost/tools:%s/\n", port)
	log.Fatal(httpServer.ListenAndServe())
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		{
			p := page.New()
			p.Title = "Generate Text to Image in Go"
			renderPage(w, r, p, "views/base.html", "views/text-to-image-builder.html")
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

			b, err := generateImage(textContent, fgColorHex, bgColorHex, fontSize)
			if err != nil {
				log.Println(err)
				return
			}

			str := base64.StdEncoding.EncodeToString(b)
			p := page.New()
			p.Title = "Generate Text to Image in Go"

			data := make(map[string]interface{})
			data["generatedImage"] = str

			p.SetData(data)
			renderPage(w, r, p, "views/base.html", "views/text-to-image-result.html")
		}
	default:
		{
		}
	}
}

func generateImage(textContent string, fgColorHex string, bgColorHex string, fontSize float64) ([]byte, error) {

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

	loadedFont, err := loadFont()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	code := strings.Replace(textContent, "\t", "    ", -1) // convert tabs into spaces
	text := strings.Split(code, "\n")                      // split newlines into arrays

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
	textYOffset := 10 + int(c.PointToFixed(fontSize)>>6) // Note shift/truncate 6 bits first

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

func renderPage(w http.ResponseWriter, r *http.Request, vm interface{}, filenames ...string) {
	page := vm.(*page.Page)

	if page.Data == nil {
		page.SetData(make(map[string]interface{}))
	}

	if page.ErrorMessages == nil {
		page.ResetErrors()
	}

	if page.UIMapData == nil {
		page.UIMapData = make(map[string]interface{})
	}

	templateFS, err := template.ParseFS(views, "views/*")
	x, err := templateFS.New("base").ParseFiles(filenames...)
	if err != nil {
		log.Panic(err.Error())
	}
	err = x.Execute(w, page)
	if err != nil {
		log.Panic(err.Error())
	}
}

func loadFont() (*truetype.Font, error) {
	fontFile = "static/fonts/UbuntuMono-R.ttf"
	fontBytes, err := os.ReadFile(fontFile)
	if err != nil {
		return nil, err
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// This will handle the loading of config.yml
func loadConfiguration(a string, b *models.Config) {
	f, err := os.Open(a)
	if err != nil {
		log.Fatal(err.Error())
	}

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(b)
	if err != nil {
		log.Fatal(err.Error())
	}
}
