package main

import (
	"bytes"
	"embed"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/golang/freetype"
	"github.com/gorilla/csrf"
	"github.com/johnpili/go-text-to-image/models"
	"github.com/johnpili/go-text-to-image/page"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/image/font"
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
	router.ServeFiles(fmt.Sprintf("%s%s", configuration.HTTP.BasePath, "/static/*filepath"), http.FS(v))
	router.HandlerFunc("GET", fmt.Sprintf("%s%s", configuration.HTTP.BasePath, "/"), indexHandler)
	router.HandlerFunc("POST", fmt.Sprintf("%s%s", configuration.HTTP.BasePath, "/"), indexHandler)

	csrfProtection := csrf.Protect(generateRandomBytes(32))
	port := strconv.Itoa(configuration.HTTP.Port)
	httpServer := &http.Server{
		Addr:         ":" + port,
		Handler:      csrfProtection(router),
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
	}

	if configuration.HTTP.IsTLS {
		log.Printf("Server running at https://localhost:%s%s/\n", port, configuration.HTTP.BasePath)
		log.Fatal(httpServer.ListenAndServeTLS(configuration.HTTP.ServerCert, configuration.HTTP.ServerKey))
		return
	}
	log.Printf("Server running at http://localhost:%s%s/\n", port, configuration.HTTP.BasePath)
	log.Fatal(httpServer.ListenAndServe())
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		{
			p := page.New()
			p.Title = "Generate Text to Image in Go"
			p.CSRFToken = csrf.Token(r)
			renderPage(w, r, p, configuration.HTTP.BasePath, "views/base.html", "views/text-to-image-builder.html")
		}
	case http.MethodPost:
		{
			textContent := strings.Trim(r.FormValue("textContent"), " ")

			fontSize, err := strconv.ParseFloat(strings.Trim(r.FormValue("fontSize"), " "), 64)
			if err != nil {
				fontSize = 32
			}

			fontName := strings.Trim(r.FormValue("fontName"), " ")

			fontColorHex := strings.Trim(r.FormValue("fontColor"), " ")
			fontColorHex = strings.ToLower(fontColorHex)

			backgroundColorHex := strings.Trim(r.FormValue("backgroundColor"), " ")
			backgroundColorHex = strings.ToLower(backgroundColorHex)

			b, err := generateImage(textContent, fontColorHex, backgroundColorHex, fontName, fontSize)
			if err != nil {
				log.Println(err)
				return
			}

			str := base64.StdEncoding.EncodeToString(b)
			p := page.New()
			p.Title = "Generate Text to Image in Go"
			p.CSRFToken = csrf.Token(r)

			data := make(map[string]interface{})
			data["generatedImage"] = str

			p.SetData(data)
			renderPage(w, r, p, configuration.HTTP.BasePath, "views/base.html", "views/text-to-image-result.html")
		}
	default:
		{
		}
	}
}

func generateImage(textContent string, fgColorHex string, bgColorHex string, fontName string, fontSize float64) ([]byte, error) {

	fgColor := color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff} // Default font color
	if len(fgColorHex) == 7 {
		_, err := fmt.Sscanf(fgColorHex, "#%02x%02x%02x", &fgColor.R, &fgColor.G, &fgColor.B)
		if err != nil {
			log.Println(err)
			fgColor = color.RGBA{R: 0x2e, G: 0x34, B: 0x36, A: 0xff}
		}
	}

	bgColor := color.RGBA{R: 0x30, G: 0x0a, B: 0x24, A: 0xff} // Default background color
	if len(bgColorHex) == 7 {
		_, err := fmt.Sscanf(bgColorHex, "#%02x%02x%02x", &bgColor.R, &bgColor.G, &bgColor.B)
		if err != nil {
			log.Println(err)
			bgColor = color.RGBA{R: 0x30, G: 0x0a, B: 0x24, A: 0xff}
		}
	}

	loadedFont, err := loadFont(fontName)
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
