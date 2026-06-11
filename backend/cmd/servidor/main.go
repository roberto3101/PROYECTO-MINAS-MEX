package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"minas/capacidades/gobierno/aplicacion"
	"minas/capacidades/gobierno/entrada"
	"minas/capacidades/gobierno/infraestructura"
	"minas/compartido/reloj"
	"minas/plataforma/entrada/web"
	"minas/plataforma/identidad"
	"minas/plataforma/persistencia"
	"minas/plataforma/seguridad"
)

func main() {
	ctx := context.Background()

	cadenaDeConexion := variableObligatoria("CADENA_POSTGRES")
	secretoDelToken := variableObligatoria("SECRETO_TOKEN")
	direccion := variableConValorPorDefecto("DIRECCION", ":8080")

	pool, err := persistencia.NuevoPool(ctx, cadenaDeConexion)
	if err != nil {
		log.Fatalf("no se pudo conectar a postgres: %v", err)
	}
	defer pool.Close()

	relojDelSistema := reloj.DelSistema()
	unidad := persistencia.NuevaUnidadDeTrabajo(pool)
	unidadDePlataforma := persistencia.NuevaUnidadDePlataforma(pool)

	repositorioUsuario := infraestructura.NuevoRepositorioUsuario()
	repositorioRol := infraestructura.NuevoRepositorioRol()
	repositorioAsignacion := infraestructura.NuevoRepositorioAsignacion()
	repositorioEmpresa := infraestructura.NuevoRepositorioEmpresa()
	repositorioPermiso := infraestructura.NuevoRepositorioPermiso()
	lectorDeAcceso := infraestructura.NuevoLectorDeAcceso()
	servicioGobierno := infraestructura.NuevoServicioGobierno(unidad)
	cifrador := seguridad.NuevoCifradorBcrypt()

	emisor := identidad.NuevoEmisorDeToken(secretoDelToken, 8*time.Hour)

	manejador := entrada.NuevoManejadorGobierno(
		aplicacion.NuevoIniciarSesion(unidadDePlataforma, lectorDeAcceso, cifrador),
		aplicacion.NuevoRegistrarUsuario(unidad, repositorioUsuario),
		aplicacion.NuevoDesactivarUsuario(unidad, repositorioUsuario),
		aplicacion.NuevoCrearRol(unidad, repositorioRol),
		aplicacion.NuevoConcederPermisoARol(unidad, repositorioRol, repositorioPermiso),
		aplicacion.NuevoAsignarRol(unidad, repositorioAsignacion),
		aplicacion.NuevoRevocarRol(unidad, repositorioAsignacion),
		aplicacion.NuevoConfigurarEmpresa(unidad, repositorioEmpresa),
		servicioGobierno,
		emisor,
		relojDelSistema,
	)

	rutas := http.NewServeMux()
	rutas.HandleFunc("GET /salud", func(escritor http.ResponseWriter, _ *http.Request) {
		web.ResponderJson(escritor, http.StatusOK, map[string]string{"estado": "vivo"})
	})
	manejador.Registrar(rutas, web.NuevoAutenticador(emisor, relojDelSistema))

	servidor := web.NuevoServidor(direccion, rutas)
	log.Printf("servidor de gobierno escuchando en %s", direccion)
	if err := servidor.ListenAndServe(); err != nil {
		log.Fatalf("el servidor se detuvo: %v", err)
	}
}

func variableObligatoria(nombre string) string {
	valor := os.Getenv(nombre)
	if valor == "" {
		log.Fatalf("falta la variable de entorno %s", nombre)
	}
	return valor
}

func variableConValorPorDefecto(nombre, porDefecto string) string {
	if valor := os.Getenv(nombre); valor != "" {
		return valor
	}
	return porDefecto
}
