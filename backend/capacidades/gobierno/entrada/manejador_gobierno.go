package entrada

import (
	"net/http"

	"minas/capacidades/gobierno/aplicacion"
	"minas/capacidades/gobierno/contrato"
	"minas/compartido/reloj"
	"minas/plataforma/contexto"
	"minas/plataforma/entrada/web"
	"minas/plataforma/identidad"
)

type ManejadorGobierno struct {
	iniciarSesion     *aplicacion.IniciarSesion
	registrarUsuario  *aplicacion.RegistrarUsuario
	desactivarUsuario *aplicacion.DesactivarUsuario
	crearRol          *aplicacion.CrearRol
	concederPermiso   *aplicacion.ConcederPermisoARol
	asignarRol        *aplicacion.AsignarRol
	revocarRol        *aplicacion.RevocarRol
	configurarEmpresa *aplicacion.ConfigurarEmpresa
	gobierno          contrato.Gobierno
	emisor            identidad.EmisorDeToken
	reloj             reloj.Reloj
}

func NuevoManejadorGobierno(
	iniciarSesion *aplicacion.IniciarSesion,
	registrarUsuario *aplicacion.RegistrarUsuario,
	desactivarUsuario *aplicacion.DesactivarUsuario,
	crearRol *aplicacion.CrearRol,
	concederPermiso *aplicacion.ConcederPermisoARol,
	asignarRol *aplicacion.AsignarRol,
	revocarRol *aplicacion.RevocarRol,
	configurarEmpresa *aplicacion.ConfigurarEmpresa,
	gobierno contrato.Gobierno,
	emisor identidad.EmisorDeToken,
	relojDelSistema reloj.Reloj,
) *ManejadorGobierno {
	return &ManejadorGobierno{
		iniciarSesion:     iniciarSesion,
		registrarUsuario:  registrarUsuario,
		desactivarUsuario: desactivarUsuario,
		crearRol:          crearRol,
		concederPermiso:   concederPermiso,
		asignarRol:        asignarRol,
		revocarRol:        revocarRol,
		configurarEmpresa: configurarEmpresa,
		gobierno:          gobierno,
		emisor:            emisor,
		reloj:             relojDelSistema,
	}
}

func (manejador *ManejadorGobierno) Registrar(rutas *http.ServeMux, autenticador web.Autenticador) {
	rutas.HandleFunc("POST /sesiones", manejador.iniciarSesionHttp)
	rutas.Handle("POST /gobierno/usuarios", autenticador.Exigir("usuarios.crear", manejador.registrarUsuarioHttp))
	rutas.Handle("DELETE /gobierno/usuarios/{id}", autenticador.Exigir("usuarios.desactivar", manejador.desactivarUsuarioHttp))
	rutas.Handle("POST /gobierno/roles", autenticador.Exigir("roles.crear", manejador.crearRolHttp))
	rutas.Handle("POST /gobierno/roles/{id}/permisos", autenticador.Exigir("roles.editar", manejador.concederPermisoHttp))
	rutas.Handle("POST /gobierno/asignaciones", autenticador.Exigir("roles.asignar", manejador.asignarRolHttp))
	rutas.Handle("DELETE /gobierno/asignaciones/{id}", autenticador.Exigir("roles.asignar", manejador.revocarRolHttp))
	rutas.Handle("PUT /gobierno/empresa", autenticador.Exigir("empresa.configurar", manejador.configurarEmpresaHttp))
	rutas.Handle("GET /gobierno/permisos-vigentes", autenticador.Requerir(http.HandlerFunc(manejador.permisosVigentesHttp)))
}

func (manejador *ManejadorGobierno) iniciarSesionHttp(escritor http.ResponseWriter, peticion *http.Request) {
	var cuerpo struct {
		CodigoEmpresa string `json:"codigo_empresa"`
		NombreCorto   string `json:"usuario"`
		Contrasena    string `json:"contrasena"`
	}
	if !web.DecodificarCuerpo(escritor, peticion, &cuerpo) {
		return
	}
	sesion, err := manejador.iniciarSesion.Ejecutar(peticion.Context(), aplicacion.ComandoIniciarSesion{
		CodigoEmpresa: cuerpo.CodigoEmpresa,
		NombreCorto:   cuerpo.NombreCorto,
		Contrasena:    cuerpo.Contrasena,
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	token, err := manejador.emisor.Emitir(identidad.Sesion{
		IdentificadorUsuario: sesion.IdentificadorUsuario,
		IdentificadorEmpresa: sesion.IdentificadorEmpresa,
		NombreCorto:          sesion.NombreCorto,
		Permisos:             sesion.Permisos,
	}, manejador.reloj.Ahora())
	if err != nil {
		web.ResponderError(escritor, http.StatusInternalServerError, "no se pudo emitir el token")
		return
	}
	web.ResponderJson(escritor, http.StatusOK, map[string]any{"token": token, "permisos": sesion.Permisos})
}

func (manejador *ManejadorGobierno) registrarUsuarioHttp(escritor http.ResponseWriter, peticion *http.Request) {
	var cuerpo struct {
		NombreCorto           string `json:"usuario"`
		Nombre                string `json:"nombre"`
		Correo                string `json:"correo"`
		IdentificadorEmpleado string `json:"id_empleado"`
	}
	if !web.DecodificarCuerpo(escritor, peticion, &cuerpo) {
		return
	}
	identificadorUsuario, err := manejador.registrarUsuario.Ejecutar(peticion.Context(), aplicacion.ComandoRegistrarUsuario{
		IdentificadorEmpresa:  empresaDe(peticion),
		NombreCorto:           cuerpo.NombreCorto,
		Nombre:                cuerpo.Nombre,
		Correo:                cuerpo.Correo,
		IdentificadorEmpleado: cuerpo.IdentificadorEmpleado,
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusCreated, map[string]string{"id": identificadorUsuario})
}

func (manejador *ManejadorGobierno) desactivarUsuarioHttp(escritor http.ResponseWriter, peticion *http.Request) {
	err := manejador.desactivarUsuario.Ejecutar(peticion.Context(), aplicacion.ComandoDesactivarUsuario{
		IdentificadorUsuario: peticion.PathValue("id"),
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, map[string]string{"estado": "INACTIVO"})
}

func (manejador *ManejadorGobierno) crearRolHttp(escritor http.ResponseWriter, peticion *http.Request) {
	var cuerpo struct {
		Codigo      string `json:"codigo"`
		Descripcion string `json:"descripcion"`
	}
	if !web.DecodificarCuerpo(escritor, peticion, &cuerpo) {
		return
	}
	identificadorRol, err := manejador.crearRol.Ejecutar(peticion.Context(), aplicacion.ComandoCrearRol{
		IdentificadorEmpresa: empresaDe(peticion),
		Codigo:               cuerpo.Codigo,
		Descripcion:          cuerpo.Descripcion,
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusCreated, map[string]string{"id": identificadorRol})
}

func (manejador *ManejadorGobierno) concederPermisoHttp(escritor http.ResponseWriter, peticion *http.Request) {
	var cuerpo struct {
		CodigoPermiso string `json:"permiso"`
	}
	if !web.DecodificarCuerpo(escritor, peticion, &cuerpo) {
		return
	}
	err := manejador.concederPermiso.Ejecutar(peticion.Context(), aplicacion.ComandoConcederPermisoARol{
		IdentificadorRol: peticion.PathValue("id"),
		CodigoPermiso:    cuerpo.CodigoPermiso,
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, map[string]string{"estado": "CONCEDIDO"})
}

func (manejador *ManejadorGobierno) asignarRolHttp(escritor http.ResponseWriter, peticion *http.Request) {
	var cuerpo struct {
		IdentificadorUsuario string `json:"id_usuario"`
		IdentificadorRol     string `json:"id_rol"`
		AlcanceMina          string `json:"id_mina"`
	}
	if !web.DecodificarCuerpo(escritor, peticion, &cuerpo) {
		return
	}
	identificadorAsignacion, err := manejador.asignarRol.Ejecutar(peticion.Context(), aplicacion.ComandoAsignarRol{
		IdentificadorEmpresa: empresaDe(peticion),
		IdentificadorUsuario: cuerpo.IdentificadorUsuario,
		IdentificadorRol:     cuerpo.IdentificadorRol,
		AlcanceMina:          cuerpo.AlcanceMina,
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusCreated, map[string]string{"id": identificadorAsignacion})
}

func (manejador *ManejadorGobierno) revocarRolHttp(escritor http.ResponseWriter, peticion *http.Request) {
	err := manejador.revocarRol.Ejecutar(peticion.Context(), aplicacion.ComandoRevocarRol{
		IdentificadorAsignacion: peticion.PathValue("id"),
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, map[string]string{"estado": "REVOCADA"})
}

func (manejador *ManejadorGobierno) configurarEmpresaHttp(escritor http.ResponseWriter, peticion *http.Request) {
	var cuerpo struct {
		LogoUrl       string `json:"logo_url"`
		ColorPrimario string `json:"color_primario"`
		ZonaHoraria   string `json:"zona_horaria"`
		Moneda        string `json:"moneda"`
	}
	if !web.DecodificarCuerpo(escritor, peticion, &cuerpo) {
		return
	}
	err := manejador.configurarEmpresa.Ejecutar(peticion.Context(), aplicacion.ComandoConfigurarEmpresa{
		IdentificadorEmpresa: empresaDe(peticion),
		LogoUrl:              cuerpo.LogoUrl,
		ColorPrimario:        cuerpo.ColorPrimario,
		ZonaHoraria:          cuerpo.ZonaHoraria,
		Moneda:               cuerpo.Moneda,
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, map[string]string{"estado": "CONFIGURADA"})
}

func (manejador *ManejadorGobierno) permisosVigentesHttp(escritor http.ResponseWriter, peticion *http.Request) {
	sesion, _ := web.SesionDe(peticion.Context())
	permisos, err := manejador.gobierno.PermisosVigentesDe(peticion.Context(), sesion.IdentificadorUsuario)
	if err != nil {
		web.ResponderError(escritor, http.StatusInternalServerError, err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, permisos)
}

func empresaDe(peticion *http.Request) string {
	tenant, _ := contexto.TenantDe(peticion.Context())
	return tenant.Empresa.Texto()
}
