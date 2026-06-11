package web

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
)

func NuevoServidorDeFrontend(directorio string) http.Handler {
	archivos := http.FileServer(http.Dir(directorio))
	paginaNoEncontrada := filepath.Join(directorio, "404.html")
	return http.HandlerFunc(func(escritor http.ResponseWriter, peticion *http.Request) {
		rutaLimpia := path.Clean(peticion.URL.Path)
		if rutaLimpia == "/" {
			archivos.ServeHTTP(escritor, peticion)
			return
		}
		destino := filepath.Join(directorio, filepath.FromSlash(rutaLimpia))
		informacion, err := os.Stat(destino)
		if err != nil || informacion.IsDir() {
			servirPaginaNoEncontrada(escritor, paginaNoEncontrada)
			return
		}
		archivos.ServeHTTP(escritor, peticion)
	})
}

func servirPaginaNoEncontrada(escritor http.ResponseWriter, ruta string) {
	contenido, err := os.ReadFile(ruta)
	if err != nil {
		http.Error(escritor, "404 - pagina no encontrada", http.StatusNotFound)
		return
	}
	escritor.Header().Set("Content-Type", "text/html; charset=utf-8")
	escritor.WriteHeader(http.StatusNotFound)
	_, _ = escritor.Write(contenido)
}
