package main

import (
	"context"
	"log"
	"os"
	"time"

	aplicacionCatalogos "minas/capacidades/catalogos/aplicacion"
	entradaCatalogos "minas/capacidades/catalogos/entrada"
	infraCatalogos "minas/capacidades/catalogos/infraestructura"
	aplicacionGobierno "minas/capacidades/gobierno/aplicacion"
	entradaGobierno "minas/capacidades/gobierno/entrada"
	infraGobierno "minas/capacidades/gobierno/infraestructura"
	"minas/compartido/reloj"
	"minas/pasarela"
	"minas/plataforma/archivos"
	"minas/plataforma/entrada/web"
	"minas/plataforma/escudo"
	"minas/plataforma/identidad"
	"minas/plataforma/persistencia"
	"minas/plataforma/seguridad"
)

func main() {
	ctx := context.Background()

	cadenaDeConexion := variableObligatoria("CADENA_POSTGRES")
	secretoDelToken := variableObligatoria("SECRETO_TOKEN")
	direccion := variableConValorPorDefecto("DIRECCION", ":8080")
	directorioFrontend := variableConValorPorDefecto("DIRECTORIO_FRONTEND", "../frontend")
	directorioArchivos := variableConValorPorDefecto("DIRECTORIO_ARCHIVOS", "../datos")

	pool, err := persistencia.NuevoPool(ctx, cadenaDeConexion)
	if err != nil {
		log.Fatalf("no se pudo conectar a postgres: %v", err)
	}
	defer pool.Close()

	relojDelSistema := reloj.DelSistema()
	unidad := persistencia.NuevaUnidadDeTrabajo(pool)
	unidadDePlataforma := persistencia.NuevaUnidadDePlataforma(pool)
	cifrador := seguridad.NuevoCifradorBcrypt()
	emisor := identidad.NuevoEmisorDeToken(secretoDelToken, 8*time.Hour)
	limitador := escudo.NuevoLimitadorDeIntentos()
	almacen := archivos.NuevoAlmacenLocal(directorioArchivos)

	repositorioUsuario := infraGobierno.NuevoRepositorioUsuario()
	repositorioRol := infraGobierno.NuevoRepositorioRol()
	repositorioAsignacion := infraGobierno.NuevoRepositorioAsignacion()
	repositorioEmpresa := infraGobierno.NuevoRepositorioEmpresa()
	repositorioPermiso := infraGobierno.NuevoRepositorioPermiso()
	lectorDeAcceso := infraGobierno.NuevoLectorDeAcceso()
	lectorDeGobierno := infraGobierno.NuevoLectorDeGobierno()
	lectorDePlataforma := infraGobierno.NuevoLectorDePlataforma()
	aprovisionadorDeAccesos := infraGobierno.NuevoAprovisionadorDeAccesos()
	servicioGobierno := infraGobierno.NuevoServicioGobierno(unidad)

	repositorioMina := infraCatalogos.NuevoRepositorioMina()
	repositorioEmpleado := infraCatalogos.NuevoRepositorioEmpleado()
	repositorioEquipo := infraCatalogos.NuevoRepositorioEquipo()
	lectorDeCatalogos := infraCatalogos.NuevoLectorDeCatalogos()
	servicioCatalogos := infraCatalogos.NuevoServicioCatalogos(unidad)

	manejadorGobierno := entradaGobierno.NuevoManejadorGobierno(
		aplicacionGobierno.NuevoIniciarSesion(unidadDePlataforma, lectorDeAcceso, cifrador, limitador, relojDelSistema),
		aplicacionGobierno.NuevoIniciarSesionDePlataforma(unidadDePlataforma, lectorDePlataforma, cifrador, limitador, relojDelSistema),
		aplicacionGobierno.NuevoAprovisionarEmpresa(unidadDePlataforma, unidad, lectorDePlataforma, repositorioEmpresa,
			repositorioUsuario, repositorioAsignacion, aprovisionadorDeAccesos, cifrador, servicioCatalogos),
		aplicacionGobierno.NuevoListarEmpresas(unidadDePlataforma, lectorDePlataforma),
		aplicacionGobierno.NuevoDetalleDeEmpresa(unidadDePlataforma, lectorDePlataforma),
		aplicacionGobierno.NuevoCambiarEstadoDeEmpresa(unidadDePlataforma, repositorioEmpresa),
		aplicacionGobierno.NuevoRegistrarUsuario(unidad, repositorioUsuario, cifrador),
		aplicacionGobierno.NuevoListarUsuarios(unidad, lectorDeGobierno),
		aplicacionGobierno.NuevoDetalleDeUsuario(unidad, lectorDeGobierno),
		aplicacionGobierno.NuevoEditarUsuario(unidad, repositorioUsuario),
		aplicacionGobierno.NuevoCambiarEstadoDeUsuario(unidad, repositorioUsuario),
		aplicacionGobierno.NuevoCrearRol(unidad, repositorioRol),
		aplicacionGobierno.NuevoListarRoles(unidad, lectorDeGobierno),
		aplicacionGobierno.NuevoListarPermisos(unidad, lectorDeGobierno),
		aplicacionGobierno.NuevoConcederPermisoARol(unidad, repositorioRol, repositorioPermiso),
		aplicacionGobierno.NuevoAsignarRol(unidad, repositorioAsignacion),
		aplicacionGobierno.NuevoRevocarRol(unidad, repositorioAsignacion),
		aplicacionGobierno.NuevoListarAsignacionesDeUsuario(unidad, lectorDeGobierno),
		aplicacionGobierno.NuevoConfigurarEmpresa(unidad, repositorioEmpresa),
		aplicacionGobierno.NuevoDefinirLogoDeEmpresa(unidad, repositorioEmpresa, almacen),
		servicioGobierno,
		emisor,
		relojDelSistema,
	)

	manejadorCatalogos := entradaCatalogos.NuevoManejadorCatalogos(
		aplicacionCatalogos.NuevoCrearMina(unidad, repositorioMina),
		aplicacionCatalogos.NuevoListarMinas(unidad, lectorDeCatalogos),
		aplicacionCatalogos.NuevoDetalleDeMina(unidad, lectorDeCatalogos),
		aplicacionCatalogos.NuevoCambiarEstadoDeMina(unidad, repositorioMina),
		aplicacionCatalogos.NuevoContratarEmpleado(unidad, repositorioEmpleado),
		aplicacionCatalogos.NuevoListarEmpleados(unidad, lectorDeCatalogos),
		aplicacionCatalogos.NuevoDetalleDeEmpleado(unidad, lectorDeCatalogos),
		aplicacionCatalogos.NuevoCambiarEstadoDeEmpleado(unidad, repositorioEmpleado),
		aplicacionCatalogos.NuevoDarDeAltaEquipo(unidad, repositorioEquipo),
		aplicacionCatalogos.NuevoListarEquipos(unidad, lectorDeCatalogos),
		aplicacionCatalogos.NuevoDetalleDeEquipo(unidad, lectorDeCatalogos),
		aplicacionCatalogos.NuevoCambiarEstadoDeEquipo(unidad, repositorioEquipo),
		aplicacionCatalogos.NuevoListarTiposDeEquipo(unidad, lectorDeCatalogos),
		aplicacionCatalogos.NuevoListarModulosDeTrabajo(unidad, lectorDeCatalogos),
		aplicacionCatalogos.NuevoListarDepartamentos(unidad, lectorDeCatalogos),
		aplicacionCatalogos.NuevoListarPuestos(unidad, lectorDeCatalogos),
		aplicacionCatalogos.NuevoListarActividades(unidad, lectorDeCatalogos),
	)

	rutas := pasarela.NuevasRutas(pasarela.Dependencias{
		Autenticador:       web.NuevoAutenticador(emisor, relojDelSistema),
		Gobierno:           manejadorGobierno,
		Catalogos:          manejadorCatalogos,
		Frontend:           web.NuevoServidorDeFrontend(directorioFrontend),
		DirectorioArchivos: directorioArchivos,
	})

	servidor := web.NuevoServidor(direccion, web.ConCabecerasDeSeguridad(rutas))
	log.Printf("servidor escuchando en %s (frontend: %s, archivos: %s)", direccion, directorioFrontend, directorioArchivos)
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
