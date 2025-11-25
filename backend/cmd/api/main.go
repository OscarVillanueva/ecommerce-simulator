package main

import (
	"fmt"
	"net/http"

	"github.com/OscarVillanueva/goapi/internal/app/handlers"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

func main()  {
	log.SetReportCaller(true)

	var r *chi.Mux = chi.NewRouter()
	
	handlers.Router(r)

	fmt.Println("Starting GO API Server")

	err := http.ListenAndServe("localhost:4321", r)

	if err != nil {
		log.Error(err)
	}
}


