package main

import (
	"github.com/lxc/lxd/shared/logger"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func StartServer(cfg *Config) {
	router := mux.NewRouter()

	router.HandleFunc("/instance/{server}/{name}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		d, err := InitLxdInstanceServer(cfg, vars["server"])
		if err != nil {
			logger.Error(err.Error())
			return
		}
		server := *d
		w.Header().Set("Access-Control-Allow-Origin", "*")
		socket := make(chan string)

		go vga(cfg, server, vars["name"], socket)
		spiceSocket := <-socket
		w.WriteHeader(302)
		_, err = w.Write([]byte(spiceSocket))
		if err != nil {
			logger.Error(err.Error())
			return
		}
	})

	srv := &http.Server{
		Handler: router,
		Addr:    "0.0.0.0:8085",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
