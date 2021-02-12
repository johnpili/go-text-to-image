package controllers

import (
	"github.com/johnpili/go-text-to-image/models"
	"log"
	"net/http"

	rice "github.com/GeertJohan/go.rice"
	"github.com/gorilla/sessions"
	"github.com/psi-incontrol/go-sprocket/page"
	"github.com/psi-incontrol/go-sprocket/sprocket"
)

var (
	cookieStore   *sessions.CookieStore
	viewBox       *rice.Box
	staticBox     *rice.Box
	configuration *models.Config
)

//New ...
func New(vb *rice.Box, sb *rice.Box, store *sessions.CookieStore, config *models.Config) *Hub {
	viewBox = vb
	staticBox = sb
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

	x, err := sprocket.GetTemplates(viewBox, filenames)
	err = x.Execute(w, page)
	if err != nil {
		log.Panic(err.Error())
	}
}
