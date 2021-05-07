package main

import (
	lxd "github.com/lxc/lxd/client"
	"github.com/lxc/lxd/shared/logger"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Server struct {
	cfg                 *Config
	SendDisconnectVMMap map[string]chan bool
}

func InitServer(cfg *Config) *Server {
	s := new(Server)
	s.cfg = cfg
	s.SendDisconnectVMMap = make(map[string]chan bool)
	return s
}

func (s *Server) handleDisconnectRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	name := vars["server"] + "-" + vars["name"]

	if val, ok := s.SendDisconnectVMMap[name]; ok {
		val <- true
	}

	w.WriteHeader(200)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		log.Fatal(err)
		return
	}
}

func (s *Server) handleProjectDisconnectRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	name := vars["server"] + "-" + vars["project"] + "-" + vars["name"]

	if val, ok := s.SendDisconnectVMMap[name]; ok {
		val <- true
	}

	w.WriteHeader(200)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		log.Fatal(err)
		return
	}
}

func (s *Server) getLxdServer(ip string) (lxd.InstanceServer, error) {
	d, err := InitLxdInstanceServer(s.cfg, ip)
	if err != nil {
		return nil, err
	}
	return *d, nil
}

func (s *Server) handleConsoleRequest(w http.ResponseWriter, r *http.Request, server lxd.InstanceServer, mapName string) {
	vars := mux.Vars(r)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	socket := make(chan string)

	if _, ok := s.SendDisconnectVMMap[mapName]; ok {
		// Delete directly now.
		// TODO: If this api could be called twice, and in this situation the previous disconnection cannot be closed.
		delete(s.SendDisconnectVMMap, mapName)
	}
	s.SendDisconnectVMMap[mapName] = make(chan bool)

	go vga(s.cfg, server, vars["name"], socket, s.SendDisconnectVMMap[mapName])
	spiceSocket := <-socket
	w.WriteHeader(200)
	_, err := w.Write([]byte(spiceSocket))
	if err != nil {
		logger.Error(err.Error())
		return
	}
}

func (s *Server) StartServer() {
	router := mux.NewRouter()

	router.HandleFunc("/disconnect/{server}/{name}", s.handleDisconnectRequest)
	router.HandleFunc("/disconnect/{server}/{name}/{project}", s.handleProjectDisconnectRequest)
	router.HandleFunc("/instance/{server}/{name}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		mapName := vars["server"] + "-" + vars["name"]

		server, err := s.getLxdServer(vars["server"])
		if err != nil {
			log.Fatal(err)
		}
		s.handleConsoleRequest(w, r, server, mapName)
	})
	router.HandleFunc("/instance/{server}/{name}/{project}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		mapName := vars["server"] + "-" + vars["project"] + "-" + vars["name"]

		server, err := s.getLxdServer(vars["server"])
		if err != nil {
			log.Fatal(err)
		}
		server = server.UseProject(vars["project"])
		s.handleConsoleRequest(w, r, server, mapName)
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
