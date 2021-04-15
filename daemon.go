package main

import (
	"fmt"
	lxd "github.com/lxc/lxd/client"
	"io/ioutil"
)

func InitLxdInstanceServer(cfg *Config, addr string) (*lxd.InstanceServer, error) {
	cert, err := ioutil.ReadFile(cfg.Server.Cert)
	if err != nil {
		return nil, err
	}

	key, err := ioutil.ReadFile(cfg.Server.Key)
	if err != nil {
		return nil, err
	}

	args := &lxd.ConnectionArgs{
		TLSClientCert:      string(cert),
		TLSClientKey:       string(key),
		InsecureSkipVerify: true,
	}
	server, err := lxd.ConnectLXD(fmt.Sprintf("https://%s:%s", addr, cfg.Server.Port), args)
	if err != nil {
		return nil, err
	}

	return &server, nil
}
