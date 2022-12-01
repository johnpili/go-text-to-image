package main

import (
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/johnpili/go-text-to-image/models"
	"github.com/johnpili/go-text-to-image/page"
	"gopkg.in/yaml.v2"
	"html/template"
	"log"
	"net/http"
	"os"
)

func renderPage(w http.ResponseWriter, r *http.Request, vm interface{}, basePath string, filenames ...string) {
	p := vm.(*page.Page)

	if p.Data == nil {
		p.SetData(make(map[string]interface{}))
	}

	if p.ErrorMessages == nil {
		p.ResetErrors()
	}

	if p.UIMapData == nil {
		p.UIMapData = make(map[string]interface{})
	}
	p.UIMapData["basePath"] = basePath

	templateFS := template.Must(template.New("base").ParseFS(views, filenames...))
	err := templateFS.Execute(w, p)
	if err != nil {
		log.Panic(err.Error())
	}
}

func loadFont() (*truetype.Font, error) {
	fontFile = "static/fonts/UbuntuMono-R.ttf"
	fontBytes, err := static.ReadFile(fontFile)
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
