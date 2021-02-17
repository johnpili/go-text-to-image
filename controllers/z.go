package controllers

import (
	"embed"
	"github.com/johnpili/go-text-to-image/models"
	"html/template"
	"log"
	"net/http"

	//rice "github.com/GeertJohan/go.rice"
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

	//templateFS, err := template.ParseFS(view, "views/*")
	//x, err := templates.New("base").ParseFiles(filenames ...)
	//x, err := sprocket.GetTemplates(viewBox, filenames)

	var x *template.Template
	for i := 0; i < len(filenames); i++ {
		buf, err := view.ReadFile(filenames[i])
		//if err != nil {
		//	panic(err)
		//}
		if err != nil {
			//return nil, err
			panic(err)
		}
		if i == 0 {
			x, err = template.New("base").Parse(string(buf))
		} else {
			x.New("content").Parse(string(buf))
		}
	}
	err := x.Execute(w, page)
	if err != nil {
		log.Panic(err.Error())
	}
}
