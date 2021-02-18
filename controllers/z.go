package controllers

import (
	"embed"
	"github.com/johnpili/go-text-to-image/models"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/psi-incontrol/go-sprocket/page"
)

var (
	cookieStore   *sessions.CookieStore
	configuration *models.Config
	view *embed.FS
	static *embed.FS
)

//New ...
func New(viewFS *embed.FS, staticFS *embed.FS, store *sessions.CookieStore, config *models.Config) *Hub {
	view = viewFS
	static = staticFS
	cookieStore = store
	configuration = config
	hub := new(Hub)
	hub.LoadControllers()
	return hub
}

// Hub ...
type Hub struct {
	Controllers []interface{}
}

func renderPage(w http.ResponseWriter, r *http.Request, vm interface{}, filenames ...string) {
	page := vm.(*page.Page)

	if page.Data == nil {
		page.SetData(make(map[string]interface{}))
	}

	if page.ErrorMessages == nil {
		page.ResetErrors("")
	}

	if page.UIMapData == nil {
		page.UIMapData = make(map[string]interface{})
	}

	templateFS, err := template.ParseFS(view, "views/*")
	x, err := templateFS.New("base").ParseFiles(filenames ...)
	if err != nil {
		log.Panic(err.Error())
	}
	err = x.Execute(w, page)
	if err != nil {
		log.Panic(err.Error())
	}
}
