package controllers

import (
	"fmt"
	"log"
	"reflect"

	"github.com/go-zoo/bone"
)

// RequestMapping this interface will handle your request mapping declaration on each controller
type RequestMapping interface {
	RequestMapping(router *bone.Mux)
}

// RequestMapping implementation code of the interface
func (z *Hub) RequestMapping(router *bone.Mux, requestMapping RequestMapping) {
	requestMapping.RequestMapping(router)
}

// BindRequestMapping this method binds your request mapping into the mux router
func (z *Hub) BindRequestMapping(router *bone.Mux) {
	log.Println("Binding RequestMapping for:")
	for _, v := range z.Controllers {
		z.RequestMapping(router, v.(RequestMapping))
		rt := reflect.TypeOf(v)
		log.Println(rt)
	}

	log.Println("Binded RequestMapping are the following: ")
	for _, v := range router.Routes {
		for _, m := range v {
			log.Println(m.Method, " : ", m.Path)
		}
	}
	fmt.Println("")
}

