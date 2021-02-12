package main

import (
	"flag"
	"github.com/johnpili/go-text-to-image/controllers"
	"github.com/johnpili/go-text-to-image/models"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	rice "github.com/GeertJohan/go.rice"
	"github.com/go-zoo/bone"
	"github.com/psi-incontrol/go-sprocket/sprocket"
)

var (
	configuration models.Config
)

func main() {
	pid := os.Getpid()
	err := ioutil.WriteFile("application.pid", []byte(strconv.Itoa(pid)), 0666)
	if err != nil {
		log.Fatal(err)
	}

	var configLocation string
	flag.StringVar(&configLocation, "config", "config.yml", "Set the location of configuration file")
	flag.Parse()

	sprocket.LoadYAML(configLocation, &configuration)

	viewBox := rice.MustFindBox("views")
	staticBox := rice.MustFindBox("static")
	controllersHub := controllers.New(viewBox, nil, nil, &configuration)
	staticFileServer := http.StripPrefix("/static/", http.FileServer(staticBox.HTTPBox()))

	router := bone.New()
	router.Get("/static/", staticFileServer)
	controllersHub.BindRequestMapping(router)

	// CODE FROM https://medium.com/@mossila/running-go-behind-iis-ce1a610116df
	port := strconv.Itoa(configuration.HTTP.Port)
	if os.Getenv("ASPNETCORE_PORT") != "" { // get enviroment variable that set by ACNM
		port = os.Getenv("ASPNETCORE_PORT")
	}

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
