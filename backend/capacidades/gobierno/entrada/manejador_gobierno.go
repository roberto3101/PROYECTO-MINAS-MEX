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
	iniciarSesion      *aplicacion.IniciarSesion
	registrarUsuario   *aplicacion.RegistrarUsuario
	desactivarUsuario  *aplicacion.DesactivarUsuario
	listarUsuarios     *aplicacion.ListarUsuarios
	crearRol           *aplicacion.CrearRol
	listarRoles        *aplicacion.ListarRoles
	listarPermisos     *aplicacion.ListarPermisos
	concederPermiso    *aplicacion.ConcederPermisoARol
	asignarRol         *aplicacion.AsignarRol
	revocarRol         *aplicacion.RevocarRol
	listarAsignaciones *aplicacion.ListarAsignacionesDeUsuario
	configurarEmpresa  *aplicacion.ConfigurarEmpresa
	gobierno           contrato.Gobierno
	emisor             identidad.EmisorDeToken
	reloj              reloj.Reloj
}

func NuevoManejadorGobierno(
	iniciarSesion *aplicacion.IniciarSesion,
	registrarUsuario *aplicacion.RegistrarUsuario,
	desactivarUsuario *aplicacion.DesactivarUsuario,
	listarUsuarios *aplicacion.ListarUsuarios,
	crearRol *aplicacion.CrearRol,
	listarRoles *aplicacion.ListarRoles,
	listarPermisos *aplicacion.ListarPermisos,
	concederPermiso *aplicacion.ConcederPermisoARol,
	asignarRol *aplicacion.AsignarRol,
	revocarRol *aplicacion.RevocarRol,
	listarAsignaciones *aplicacion.ListarAsignacionesDeUsuario,
	configurarEmpresa *aplicacion.ConfigurarEmpresa,
	gobierno contrato.Gobierno,
	emisor identidad.EmisorDeToken,
	relojDelSistema reloj.Reloj,
) *ManejadorGobierno {
	return &ManejadorGobierno{
		iniciarSesion:      iniciarSesion,
		registrarUsuario:   registrarUsuario,
		desactivarUsuario:  desactivarUsuario,
		listarUsuarios:     listarUsuarios,
		crearRol:           crearRol,
		listarRoles:        listarRoles,
		listarPermisos:     listarPermisos,
		concederPermiso:    concederPermiso,
		asignarRol:         asignarRol,
		revocarRol:         revocarRol,
		listarAsignaciones: listarAsignaciones,
		configurarEmpresa:  configurarEmpresa,
		gobierno:           gobierno,
		emisor:             emisor,
		reloj:              relojDelSistema,
	}
}

func (manejador *ManejadorGobierno) IniciarSesion(escritor http.ResponseWriter, peticion *http.Request) {
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
	web.ResponderJson(escritor, http.StatusOK, map[string]any{
		"token":    token,
		"usuario":  sesion.NombreCorto,
		"permisos": sesion.Permisos,
	})
}

func (manejador *ManejadorGobierno) RegistrarUsuario(escritor http.ResponseWriter, peticion *http.Request) {
	var cuerpo struct {
		NombreCorto           string `json:"usuario"`
		Nombre                string `json:"nombre"`
		Correo                string `json:"correo"`
		Contrasena            string `json:"contrasena"`
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
		Contrasena:            cuerpo.Contrasena,
		IdentificadorEmpleado: cuerpo.IdentificadorEmpleado,
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusCreated, map[string]string{"id": identificadorUsuario})
}

func (manejador *ManejadorGobierno) DesactivarUsuario(escritor http.ResponseWriter, peticion *http.Request) {
	err := manejador.desactivarUsuario.Ejecutar(peticion.Context(), aplicacion.ComandoDesactivarUsuario{
		IdentificadorUsuario: peticion.PathValue("id"),
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, map[string]string{"estado": "INACTIVO"})
}

func (manejador *ManejadorGobierno) ListarUsuarios(escritor http.ResponseWriter, peticion *http.Request) {
	usuarios, err := manejador.listarUsuarios.Ejecutar(peticion.Context())
	if err != nil {
		web.ResponderError(escritor, http.StatusInternalServerError, err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, usuarios)
}

func (manejador *ManejadorGobierno) CrearRol(escritor http.ResponseWriter, peticion *http.Request) {
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

func (manejador *ManejadorGobierno) ListarRoles(escritor http.ResponseWriter, peticion *http.Request) {
	roles, err := manejador.listarRoles.Ejecutar(peticion.Context())
	if err != nil {
		web.ResponderError(escritor, http.StatusInternalServerError, err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, roles)
}

func (manejador *ManejadorGobierno) ListarPermisos(escritor http.ResponseWriter, peticion *http.Request) {
	permisos, err := manejador.listarPermisos.Ejecutar(peticion.Context())
	if err != nil {
		web.ResponderError(escritor, http.StatusInternalServerError, err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, permisos)
}

func (manejador *ManejadorGobierno) ConcederPermiso(escritor http.ResponseWriter, peticion *http.Request) {
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

func (manejador *ManejadorGobierno) AsignarRol(escritor http.ResponseWriter, peticion *http.Request) {
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

func (manejador *ManejadorGobierno) RevocarRol(escritor http.ResponseWriter, peticion *http.Request) {
	err := manejador.revocarRol.Ejecutar(peticion.Context(), aplicacion.ComandoRevocarRol{
		IdentificadorAsignacion: peticion.PathValue("id"),
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, map[string]string{"estado": "REVOCADA"})
}

func (manejador *ManejadorGobierno) ListarAsignacionesDeUsuario(escritor http.ResponseWriter, peticion *http.Request) {
	asignaciones, err := manejador.listarAsignaciones.Ejecutar(peticion.Context(), peticion.PathValue("id"))
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, asignaciones)
}

func (manejador *ManejadorGobierno) ConfigurarEmpresa(escritor http.ResponseWriter, peticion *http.Request) {
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

func (manejador *ManejadorGobierno) EmpresaActual(escritor http.ResponseWriter, peticion *http.Request) {
	empresa, encontrada, err := manejador.gobierno.EmpresaActual(peticion.Context())
	if err != nil {
		web.ResponderError(escritor, http.StatusInternalServerError, err.Error())
		return
	}
	if !encontrada {
		web.ResponderError(escritor, http.StatusNotFound, "empresa no encontrada")
		return
	}
	web.ResponderJson(escritor, http.StatusOK, empresa)
}

func (manejador *ManejadorGobierno) PermisosVigentes(escritor http.ResponseWriter, peticion *http.Request) {
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
