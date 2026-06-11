package pasarela

import (
	"net/http"

	entradaCatalogos "minas/capacidades/catalogos/entrada"
	entradaGobierno "minas/capacidades/gobierno/entrada"
	"minas/plataforma/entrada/web"
)

type Dependencias struct {
	Autenticador web.Autenticador
	Gobierno     *entradaGobierno.ManejadorGobierno
	Catalogos    *entradaCatalogos.ManejadorCatalogos
	Frontend     http.Handler
}

func NuevasRutas(dependencias Dependencias) *http.ServeMux {
	exigir := dependencias.Autenticador.Exigir
	autenticada := func(manejador http.HandlerFunc) http.Handler {
		return dependencias.Autenticador.Requerir(manejador)
	}
	gobierno := dependencias.Gobierno
	catalogos := dependencias.Catalogos

	rutas := http.NewServeMux()

	rutas.HandleFunc("GET /salud", estadoDelServicio)
	rutas.HandleFunc("POST /sesiones", gobierno.IniciarSesion)

	rutas.Handle("GET /gobierno/empresa", autenticada(gobierno.EmpresaActual))
	rutas.Handle("PUT /gobierno/empresa", exigir("empresa.configurar", gobierno.ConfigurarEmpresa))
	rutas.Handle("GET /gobierno/sesion/permisos", autenticada(gobierno.PermisosVigentes))

	rutas.Handle("GET /gobierno/usuarios", exigir("usuarios.ver", gobierno.ListarUsuarios))
	rutas.Handle("POST /gobierno/usuarios", exigir("usuarios.crear", gobierno.RegistrarUsuario))
	rutas.Handle("DELETE /gobierno/usuarios/{id}", exigir("usuarios.desactivar", gobierno.DesactivarUsuario))
	rutas.Handle("GET /gobierno/usuarios/{id}/asignaciones", exigir("roles.ver", gobierno.ListarAsignacionesDeUsuario))

	rutas.Handle("GET /gobierno/roles", exigir("roles.ver", gobierno.ListarRoles))
	rutas.Handle("POST /gobierno/roles", exigir("roles.crear", gobierno.CrearRol))
	rutas.Handle("POST /gobierno/roles/{id}/permisos", exigir("roles.editar", gobierno.ConcederPermiso))
	rutas.Handle("GET /gobierno/permisos", exigir("roles.ver", gobierno.ListarPermisos))

	rutas.Handle("POST /gobierno/asignaciones", exigir("roles.asignar", gobierno.AsignarRol))
	rutas.Handle("DELETE /gobierno/asignaciones/{id}", exigir("roles.asignar", gobierno.RevocarRol))

	rutas.Handle("GET /catalogos/minas", exigir("catalogos.ver", catalogos.ListarMinas))
	rutas.Handle("POST /catalogos/minas", exigir("catalogos.editar", catalogos.CrearMina))
	rutas.Handle("GET /catalogos/empleados", exigir("catalogos.ver", catalogos.ListarEmpleados))
	rutas.Handle("POST /catalogos/empleados", exigir("catalogos.editar", catalogos.ContratarEmpleado))
	rutas.Handle("GET /catalogos/equipos", exigir("catalogos.ver", catalogos.ListarEquipos))
	rutas.Handle("POST /catalogos/equipos", exigir("catalogos.editar", catalogos.DarDeAltaEquipo))
	rutas.Handle("GET /catalogos/tipos-de-equipo", exigir("catalogos.ver", catalogos.ListarTiposDeEquipo))
	rutas.Handle("GET /catalogos/modulos-de-trabajo", exigir("catalogos.ver", catalogos.ListarModulosDeTrabajo))

	if dependencias.Frontend != nil {
		rutas.Handle("GET /", dependencias.Frontend)
	}

	return rutas
}

func estadoDelServicio(escritor http.ResponseWriter, _ *http.Request) {
	web.ResponderJson(escritor, http.StatusOK, map[string]string{"estado": "vivo"})
}
