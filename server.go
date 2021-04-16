package main

import (
	"github.com/lxc/lxd/shared/logger"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Server struct {
	SendDisconnectVMMap map[string]chan bool
}

func InitServer() *Server {
	s := new(Server)
	s.SendDisconnectVMMap = make(map[string]chan bool)
	return s
}

func (s *Server) StartServer(cfg *Config) {
	router := mux.NewRouter()

	router.HandleFunc("/disconnect/{server}/{name}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		name := vars["server"] + "-" + vars["name"]

		if val, ok := s.SendDisconnectVMMap[name]; ok {
			val <- true
		}

		w.WriteHeader(200)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			logger.Error(err.Error())
			return
		}
	})

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

		mapName := vars["server"] + "-" + vars["name"]
		if _, ok := s.SendDisconnectVMMap[mapName]; ok {
			// Delete directly now.
			// TODO: If this api could be called twice, and in this situation the previous disconnection cannot be closed.
			delete(s.SendDisconnectVMMap, mapName)
		}
		s.SendDisconnectVMMap[mapName] = make(chan bool)

		go vga(cfg, server, vars["name"], socket, s.SendDisconnectVMMap[mapName])
		spiceSocket := <-socket
		w.WriteHeader(200)
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
