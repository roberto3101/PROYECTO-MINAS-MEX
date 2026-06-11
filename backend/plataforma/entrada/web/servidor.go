package web

import (
	"net/http"
	"time"
)

func NuevoServidor(direccion string, manejador http.Handler) *http.Server {
	return &http.Server{
		Addr:              direccion,
		Handler:           manejador,
		ReadHeaderTimeout: 10 * time.Second,
	}
}
