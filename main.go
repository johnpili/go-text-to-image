package main

import (
	"embed"
	"flag"
	"github.com/go-zoo/bone"
	"github.com/johnpili/go-text-to-image/controllers"
	"github.com/johnpili/go-text-to-image/models"
	"github.com/psi-incontrol/go-sprocket/sprocket"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	configuration models.Config

	//go:embed static/*
	static embed.FS

	//go:embed views/*
	views embed.FS
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

	//var x *template.Template
	//viewBox := rice.MustFindBox("views")
	//staticBox := rice.MustFindBox("static")

	//	tmp, err := viewbox.String(filenames[i])
	//	if err != nil {
	//return nil, err
	//}

	//	if i == 0 {
	//		x, err = template.New("base").Parse(tmp)
	//	} else {
	//		x.New("content").Parse(tmp)
	//	}

	controllersHub := controllers.New(&views, &static, nil, &configuration)
	//staticFileServer := http.StripPrefix("/static/", )

	router := bone.New()
	router.Get("/static/", http.FileServer(http.FS(static)))
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
