package pasarela

import (
	"net/http"

	entradaCatalogos "minas/capacidades/catalogos/entrada"
	entradaGobierno "minas/capacidades/gobierno/entrada"
	entradaProduccion "minas/capacidades/produccion/entrada"
	"minas/plataforma/entrada/web"
)

type Dependencias struct {
	Autenticador       web.Autenticador
	Gobierno           *entradaGobierno.ManejadorGobierno
	Catalogos          *entradaCatalogos.ManejadorCatalogos
	Produccion         *entradaProduccion.ManejadorProduccion
	Frontend           http.Handler
	DirectorioArchivos string
}

func NuevasRutas(dependencias Dependencias) *http.ServeMux {
	exigir := dependencias.Autenticador.Exigir
	autenticada := func(manejador http.HandlerFunc) http.Handler {
		return dependencias.Autenticador.Requerir(manejador)
	}
	plataforma := dependencias.Autenticador.RequerirPlataforma
	gobierno := dependencias.Gobierno
	catalogos := dependencias.Catalogos
	produccion := dependencias.Produccion

	rutas := http.NewServeMux()

	rutas.HandleFunc("GET /salud", estadoDelServicio)
	rutas.HandleFunc("POST /sesiones", gobierno.IniciarSesion)
	rutas.HandleFunc("POST /plataforma/sesiones", gobierno.IniciarSesionDePlataforma)

	rutas.Handle("GET /plataforma/empresas", plataforma(gobierno.ListarEmpresas))
	rutas.Handle("POST /plataforma/empresas", plataforma(gobierno.AprovisionarEmpresa))
	rutas.Handle("GET /plataforma/empresas/{id}", plataforma(gobierno.DetalleDeEmpresaDePlataforma))
	rutas.Handle("PATCH /plataforma/empresas/{id}/estado", plataforma(gobierno.CambiarEstadoDeEmpresa))
	rutas.Handle("GET /plataforma/empresas/{id}/usuarios", plataforma(gobierno.UsuariosDeEmpresa))
	rutas.Handle("GET /plataforma/empresas/{id}/roles", plataforma(gobierno.RolesDeEmpresa))
	rutas.Handle("GET /plataforma/empresas/{id}/minas", plataforma(catalogos.MinasDeEmpresa))
	rutas.Handle("GET /plataforma/empresas/{id}/usuarios/{idUsuario}/asignaciones", plataforma(gobierno.AsignacionesDeUsuarioEnEmpresa))
	rutas.Handle("POST /plataforma/empresas/{id}/asignaciones", plataforma(gobierno.AsignarRolEnEmpresa))
	rutas.Handle("DELETE /plataforma/empresas/{id}/asignaciones/{idAsignacion}", plataforma(gobierno.RevocarRolEnEmpresa))

	rutas.Handle("GET /gobierno/empresa", autenticada(gobierno.EmpresaActual))
	rutas.Handle("PUT /gobierno/empresa", exigir("empresa.configurar", gobierno.ConfigurarEmpresa))
	rutas.Handle("POST /gobierno/empresa/logo", exigir("empresa.configurar", gobierno.SubirLogo))
	rutas.Handle("GET /gobierno/sesion/permisos", autenticada(gobierno.PermisosVigentes))
	rutas.Handle("GET /gobierno/sesion/minas", autenticada(gobierno.MinasDeSesion))

	rutas.Handle("GET /gobierno/usuarios", exigir("usuarios.ver", gobierno.ListarUsuarios))
	rutas.Handle("POST /gobierno/usuarios", exigir("usuarios.crear", gobierno.RegistrarUsuario))
	rutas.Handle("GET /gobierno/usuarios/{id}", exigir("usuarios.ver", gobierno.DetalleDeUsuario))
	rutas.Handle("PUT /gobierno/usuarios/{id}", exigir("usuarios.editar", gobierno.EditarUsuario))
	rutas.Handle("PATCH /gobierno/usuarios/{id}/estado", exigir("usuarios.desactivar", gobierno.CambiarEstadoDeUsuario))
	rutas.Handle("GET /gobierno/usuarios/{id}/asignaciones", exigir("roles.ver", gobierno.ListarAsignacionesDeUsuario))

	rutas.Handle("GET /gobierno/roles", exigir("roles.ver", gobierno.ListarRoles))
	rutas.Handle("POST /gobierno/roles", exigir("roles.crear", gobierno.CrearRol))
	rutas.Handle("POST /gobierno/roles/{id}/permisos", exigir("roles.editar", gobierno.ConcederPermiso))
	rutas.Handle("GET /gobierno/permisos", exigir("roles.ver", gobierno.ListarPermisos))

	rutas.Handle("POST /gobierno/asignaciones", exigir("roles.asignar", gobierno.AsignarRol))
	rutas.Handle("DELETE /gobierno/asignaciones/{id}", exigir("roles.asignar", gobierno.RevocarRol))

	rutas.Handle("GET /catalogos/minas", exigir("catalogos.ver", catalogos.ListarMinas))
	rutas.Handle("POST /catalogos/minas", exigir("catalogos.editar", catalogos.CrearMina))
	rutas.Handle("GET /catalogos/minas/{id}", exigir("catalogos.ver", catalogos.DetalleDeMina))
	rutas.Handle("PATCH /catalogos/minas/{id}/estado", exigir("catalogos.editar", catalogos.CambiarEstadoDeMina))

	rutas.Handle("GET /catalogos/empleados", exigir("catalogos.ver", catalogos.ListarEmpleados))
	rutas.Handle("POST /catalogos/empleados", exigir("catalogos.editar", catalogos.ContratarEmpleado))
	rutas.Handle("GET /catalogos/empleados/{id}", exigir("catalogos.ver", catalogos.DetalleDeEmpleado))
	rutas.Handle("PATCH /catalogos/empleados/{id}/estado", exigir("catalogos.editar", catalogos.CambiarEstadoDeEmpleado))

	rutas.Handle("GET /catalogos/equipos", exigir("catalogos.ver", catalogos.ListarEquipos))
	rutas.Handle("POST /catalogos/equipos", exigir("catalogos.editar", catalogos.DarDeAltaEquipo))
	rutas.Handle("GET /catalogos/equipos/{id}", exigir("catalogos.ver", catalogos.DetalleDeEquipo))
	rutas.Handle("PATCH /catalogos/equipos/{id}/estado", exigir("catalogos.editar", catalogos.CambiarEstadoDeEquipo))

	rutas.Handle("GET /catalogos/obras", exigir("catalogos.ver", catalogos.ListarObras))
	rutas.Handle("POST /catalogos/obras", exigir("catalogos.editar", catalogos.CrearObra))
	rutas.Handle("GET /catalogos/obras/{id}", exigir("catalogos.ver", catalogos.DetalleDeObra))
	rutas.Handle("PATCH /catalogos/obras/{id}/estado", exigir("catalogos.editar", catalogos.CambiarEstadoDeObra))

	rutas.Handle("GET /catalogos/tipos-de-obra", exigir("catalogos.ver", catalogos.ListarTiposDeObra))
	rutas.Handle("GET /catalogos/tipos-de-mineral", exigir("catalogos.ver", catalogos.ListarTiposDeMineral))
	rutas.Handle("GET /catalogos/tipos-de-barreno", exigir("catalogos.ver", catalogos.ListarTiposDeBarreno))
	rutas.Handle("GET /catalogos/tipos-de-demora", exigir("catalogos.ver", catalogos.ListarTiposDeDemora))

	rutas.Handle("GET /produccion/indicadores", exigir("produccion.ver", produccion.Indicadores))

	rutas.Handle("GET /produccion/partes-de-acarreo", exigir("produccion.ver", produccion.ListarPartesDeAcarreo))
	rutas.Handle("POST /produccion/partes-de-acarreo", exigir("produccion.capturar", produccion.RegistrarParteDeAcarreo))
	rutas.Handle("GET /produccion/partes-de-acarreo/{id}", exigir("produccion.ver", produccion.DetalleDeParteDeAcarreo))

	rutas.Handle("GET /produccion/partes-de-rezagado", exigir("produccion.ver", produccion.ListarPartesDeRezagado))
	rutas.Handle("POST /produccion/partes-de-rezagado", exigir("produccion.capturar", produccion.RegistrarParteDeRezagado))
	rutas.Handle("GET /produccion/partes-de-rezagado/{id}", exigir("produccion.ver", produccion.DetalleDeParteDeRezagado))

	rutas.Handle("GET /produccion/partes-de-barrenacion", exigir("produccion.ver", produccion.ListarPartesDeBarrenacion))
	rutas.Handle("POST /produccion/partes-de-barrenacion", exigir("produccion.capturar", produccion.RegistrarParteDeBarrenacion))
	rutas.Handle("GET /produccion/partes-de-barrenacion/{id}", exigir("produccion.ver", produccion.DetalleDeParteDeBarrenacion))

	rutas.Handle("GET /produccion/demoras", exigir("produccion.ver", produccion.ListarDemoras))
	rutas.Handle("POST /produccion/demoras", exigir("produccion.capturar", produccion.RegistrarDemora))

	rutas.Handle("GET /catalogos/tipos-de-equipo", exigir("catalogos.ver", catalogos.ListarTiposDeEquipo))
	rutas.Handle("GET /catalogos/modulos-de-trabajo", exigir("catalogos.ver", catalogos.ListarModulosDeTrabajo))
	rutas.Handle("GET /catalogos/departamentos", exigir("catalogos.ver", catalogos.ListarDepartamentos))
	rutas.Handle("GET /catalogos/puestos", exigir("catalogos.ver", catalogos.ListarPuestos))
	rutas.Handle("GET /catalogos/actividades", exigir("catalogos.ver", catalogos.ListarActividades))

	if dependencias.DirectorioArchivos != "" {
		rutas.Handle("GET /archivos/", http.StripPrefix("/archivos/",
			http.FileServer(http.Dir(dependencias.DirectorioArchivos))))
	}
	if dependencias.Frontend != nil {
		rutas.Handle("GET /", dependencias.Frontend)
	}

	return rutas
}

func estadoDelServicio(escritor http.ResponseWriter, _ *http.Request) {
	web.ResponderJson(escritor, http.StatusOK, map[string]string{"estado": "vivo"})
}
