package entrada

import (
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"strings"

	"minas/capacidades/gobierno/aplicacion"
	"minas/capacidades/gobierno/contrato"
	"minas/capacidades/gobierno/puertos"
	"minas/compartido/paginacion"
	"minas/compartido/reloj"
	"minas/plataforma/contexto"
	"minas/plataforma/entrada/web"
	"minas/plataforma/identidad"
)

const tamanoMaximoDeLogo = 3 << 20

type ManejadorGobierno struct {
	iniciarSesion           *aplicacion.IniciarSesion
	iniciarSesionPlataforma *aplicacion.IniciarSesionDePlataforma
	aprovisionarEmpresa     *aplicacion.AprovisionarEmpresa
	listarEmpresas          *aplicacion.ListarEmpresas
	detalleEmpresa          *aplicacion.DetalleDeEmpresa
	cambiarEstadoEmpresa    *aplicacion.CambiarEstadoDeEmpresa
	registrarUsuario        *aplicacion.RegistrarUsuario
	listarUsuarios          *aplicacion.ListarUsuarios
	detalleUsuario          *aplicacion.DetalleDeUsuario
	editarUsuario           *aplicacion.EditarUsuario
	cambiarEstadoUsuario    *aplicacion.CambiarEstadoDeUsuario
	crearRol                *aplicacion.CrearRol
	listarRoles             *aplicacion.ListarRoles
	listarPermisos          *aplicacion.ListarPermisos
	concederPermiso         *aplicacion.ConcederPermisoARol
	asignarRol              *aplicacion.AsignarRol
	revocarRol              *aplicacion.RevocarRol
	listarAsignaciones      *aplicacion.ListarAsignacionesDeUsuario
	minasDeSesion           *aplicacion.MinasDeSesion
	configurarEmpresa       *aplicacion.ConfigurarEmpresa
	definirLogo             *aplicacion.DefinirLogoDeEmpresa
	gobierno                contrato.Gobierno
	emisor                  identidad.EmisorDeToken
	reloj                   reloj.Reloj
}

func NuevoManejadorGobierno(
	iniciarSesion *aplicacion.IniciarSesion,
	iniciarSesionPlataforma *aplicacion.IniciarSesionDePlataforma,
	aprovisionarEmpresa *aplicacion.AprovisionarEmpresa,
	listarEmpresas *aplicacion.ListarEmpresas,
	detalleEmpresa *aplicacion.DetalleDeEmpresa,
	cambiarEstadoEmpresa *aplicacion.CambiarEstadoDeEmpresa,
	registrarUsuario *aplicacion.RegistrarUsuario,
	listarUsuarios *aplicacion.ListarUsuarios,
	detalleUsuario *aplicacion.DetalleDeUsuario,
	editarUsuario *aplicacion.EditarUsuario,
	cambiarEstadoUsuario *aplicacion.CambiarEstadoDeUsuario,
	crearRol *aplicacion.CrearRol,
	listarRoles *aplicacion.ListarRoles,
	listarPermisos *aplicacion.ListarPermisos,
	concederPermiso *aplicacion.ConcederPermisoARol,
	asignarRol *aplicacion.AsignarRol,
	revocarRol *aplicacion.RevocarRol,
	listarAsignaciones *aplicacion.ListarAsignacionesDeUsuario,
	minasDeSesion *aplicacion.MinasDeSesion,
	configurarEmpresa *aplicacion.ConfigurarEmpresa,
	definirLogo *aplicacion.DefinirLogoDeEmpresa,
	gobierno contrato.Gobierno,
	emisor identidad.EmisorDeToken,
	relojDelSistema reloj.Reloj,
) *ManejadorGobierno {
	return &ManejadorGobierno{
		iniciarSesion:           iniciarSesion,
		iniciarSesionPlataforma: iniciarSesionPlataforma,
		aprovisionarEmpresa:     aprovisionarEmpresa,
		listarEmpresas:          listarEmpresas,
		detalleEmpresa:          detalleEmpresa,
		cambiarEstadoEmpresa:    cambiarEstadoEmpresa,
		registrarUsuario:        registrarUsuario,
		listarUsuarios:          listarUsuarios,
		detalleUsuario:          detalleUsuario,
		editarUsuario:           editarUsuario,
		cambiarEstadoUsuario:    cambiarEstadoUsuario,
		crearRol:                crearRol,
		listarRoles:             listarRoles,
		listarPermisos:          listarPermisos,
		concederPermiso:         concederPermiso,
		asignarRol:              asignarRol,
		revocarRol:              revocarRol,
		listarAsignaciones:      listarAsignaciones,
		minasDeSesion:           minasDeSesion,
		configurarEmpresa:       configurarEmpresa,
		definirLogo:             definirLogo,
		gobierno:                gobierno,
		emisor:                  emisor,
		reloj:                   relojDelSistema,
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
		CodigoEmpresa:     strings.ToUpper(strings.TrimSpace(cuerpo.CodigoEmpresa)),
		NombreCorto:       strings.TrimSpace(cuerpo.NombreCorto),
		Contrasena:        cuerpo.Contrasena,
		ClaveDeLimitacion: direccionRemota(peticion) + "|" + cuerpo.NombreCorto,
	})
	if err != nil {
		responderErrorDeSesion(escritor, err)
		return
	}
	token, err := manejador.emisor.Emitir(identidad.Sesion{
		IdentificadorUsuario: sesion.IdentificadorUsuario,
		IdentificadorEmpresa: sesion.IdentificadorEmpresa,
		NombreCorto:          sesion.NombreCorto,
		Ambito:               identidad.AmbitoEmpresa,
		Permisos:             sesion.Permisos,
	}, manejador.reloj.Ahora())
	if err != nil {
		web.ResponderError(escritor, http.StatusInternalServerError, "no se pudo emitir el token")
		return
	}
	web.ResponderJson(escritor, http.StatusOK, map[string]any{
		"token": token, "usuario": sesion.NombreCorto, "ambito": identidad.AmbitoEmpresa, "permisos": sesion.Permisos,
	})
}

func (manejador *ManejadorGobierno) IniciarSesionDePlataforma(escritor http.ResponseWriter, peticion *http.Request) {
	var cuerpo struct {
		NombreCorto string `json:"usuario"`
		Contrasena  string `json:"contrasena"`
	}
	if !web.DecodificarCuerpo(escritor, peticion, &cuerpo) {
		return
	}
	sesion, err := manejador.iniciarSesionPlataforma.Ejecutar(peticion.Context(), aplicacion.ComandoIniciarSesionDePlataforma{
		NombreCorto:       strings.TrimSpace(cuerpo.NombreCorto),
		Contrasena:        cuerpo.Contrasena,
		ClaveDeLimitacion: "plataforma|" + direccionRemota(peticion) + "|" + cuerpo.NombreCorto,
	})
	if err != nil {
		responderErrorDeSesion(escritor, err)
		return
	}
	token, err := manejador.emisor.Emitir(identidad.Sesion{
		IdentificadorUsuario: sesion.IdentificadorSuperadmin,
		NombreCorto:          sesion.NombreCorto,
		Ambito:               identidad.AmbitoPlataforma,
	}, manejador.reloj.Ahora())
	if err != nil {
		web.ResponderError(escritor, http.StatusInternalServerError, "no se pudo emitir el token")
		return
	}
	web.ResponderJson(escritor, http.StatusOK, map[string]any{
		"token": token, "usuario": sesion.NombreCorto, "ambito": identidad.AmbitoPlataforma,
	})
}

func (manejador *ManejadorGobierno) AprovisionarEmpresa(escritor http.ResponseWriter, peticion *http.Request) {
	var cuerpo struct {
		Codigo               string `json:"codigo"`
		RazonSocial          string `json:"razon_social"`
		IdentificacionFiscal string `json:"identificacion_fiscal"`
		CorreoContacto       string `json:"correo_contacto"`
		Telefono             string `json:"telefono"`
		ZonaHoraria          string `json:"zona_horaria"`
		Moneda               string `json:"moneda"`
		ColorPrimario        string `json:"color_primario"`
		Admin                struct {
			Usuario    string `json:"usuario"`
			Nombre     string `json:"nombre"`
			Correo     string `json:"correo"`
			Contrasena string `json:"contrasena"`
		} `json:"admin"`
	}
	if !web.DecodificarCuerpo(escritor, peticion, &cuerpo) {
		return
	}
	resultado, err := manejador.aprovisionarEmpresa.Ejecutar(peticion.Context(), aplicacion.ComandoAprovisionarEmpresa{
		Codigo:               cuerpo.Codigo,
		RazonSocial:          cuerpo.RazonSocial,
		IdentificacionFiscal: cuerpo.IdentificacionFiscal,
		CorreoContacto:       cuerpo.CorreoContacto,
		Telefono:             cuerpo.Telefono,
		ZonaHoraria:          cuerpo.ZonaHoraria,
		Moneda:               cuerpo.Moneda,
		ColorPrimario:        cuerpo.ColorPrimario,
		AdminUsuario:         cuerpo.Admin.Usuario,
		AdminNombre:          cuerpo.Admin.Nombre,
		AdminCorreo:          cuerpo.Admin.Correo,
		AdminContrasena:      cuerpo.Admin.Contrasena,
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusCreated, map[string]string{
		"id": resultado.IdentificadorEmpresa, "id_admin": resultado.IdentificadorAdmin,
	})
}

func (manejador *ManejadorGobierno) ListarEmpresas(escritor http.ResponseWriter, peticion *http.Request) {
	parametros := peticion.URL.Query()
	empresas, siguiente, err := manejador.listarEmpresas.Ejecutar(peticion.Context(), puertos.FiltroDeEmpresas{
		Busqueda: strings.TrimSpace(parametros.Get("busqueda")),
		Estado:   strings.ToUpper(strings.TrimSpace(parametros.Get("estado"))),
		Cursor:   parametros.Get("cursor"),
		Limite:   paginacion.LimiteSeguro(parametros.Get("limite")),
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, map[string]any{"Elementos": empresas, "SiguienteCursor": siguiente})
}

func (manejador *ManejadorGobierno) DetalleDeEmpresaDePlataforma(escritor http.ResponseWriter, peticion *http.Request) {
	detalle, encontrada, err := manejador.detalleEmpresa.Ejecutar(peticion.Context(), peticion.PathValue("id"))
	if err != nil {
		web.ResponderError(escritor, http.StatusInternalServerError, err.Error())
		return
	}
	if !encontrada {
		web.ResponderError(escritor, http.StatusNotFound, "empresa no encontrada")
		return
	}
	web.ResponderJson(escritor, http.StatusOK, detalle)
}

func (manejador *ManejadorGobierno) CambiarEstadoDeEmpresa(escritor http.ResponseWriter, peticion *http.Request) {
	var cuerpo struct {
		Estado string `json:"estado"`
	}
	if !web.DecodificarCuerpo(escritor, peticion, &cuerpo) {
		return
	}
	err := manejador.cambiarEstadoEmpresa.Ejecutar(peticion.Context(), aplicacion.ComandoCambiarEstadoDeEmpresa{
		IdentificadorEmpresa: peticion.PathValue("id"),
		Estado:               strings.ToUpper(strings.TrimSpace(cuerpo.Estado)),
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, map[string]string{"estado": strings.ToUpper(cuerpo.Estado)})
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
		NombreCorto:           strings.TrimSpace(cuerpo.NombreCorto),
		Nombre:                cuerpo.Nombre,
		Correo:                strings.TrimSpace(cuerpo.Correo),
		Contrasena:            cuerpo.Contrasena,
		IdentificadorEmpleado: cuerpo.IdentificadorEmpleado,
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusCreated, map[string]string{"id": identificadorUsuario})
}

func (manejador *ManejadorGobierno) ListarUsuarios(escritor http.ResponseWriter, peticion *http.Request) {
	parametros := peticion.URL.Query()
	usuarios, siguiente, err := manejador.listarUsuarios.Ejecutar(peticion.Context(), puertos.FiltroDeUsuarios{
		Busqueda: strings.TrimSpace(parametros.Get("busqueda")),
		Estado:   strings.ToUpper(strings.TrimSpace(parametros.Get("estado"))),
		Cursor:   parametros.Get("cursor"),
		Limite:   paginacion.LimiteSeguro(parametros.Get("limite")),
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, map[string]any{"Elementos": usuarios, "SiguienteCursor": siguiente})
}

func (manejador *ManejadorGobierno) DetalleDeUsuario(escritor http.ResponseWriter, peticion *http.Request) {
	ficha, encontrado, err := manejador.detalleUsuario.Ejecutar(peticion.Context(), peticion.PathValue("id"))
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	if !encontrado {
		web.ResponderError(escritor, http.StatusNotFound, "usuario no encontrado")
		return
	}
	web.ResponderJson(escritor, http.StatusOK, ficha)
}

func (manejador *ManejadorGobierno) EditarUsuario(escritor http.ResponseWriter, peticion *http.Request) {
	var cuerpo struct {
		Nombre string `json:"nombre"`
		Correo string `json:"correo"`
	}
	if !web.DecodificarCuerpo(escritor, peticion, &cuerpo) {
		return
	}
	err := manejador.editarUsuario.Ejecutar(peticion.Context(), aplicacion.ComandoEditarUsuario{
		IdentificadorUsuario: peticion.PathValue("id"),
		Nombre:               cuerpo.Nombre,
		Correo:               strings.TrimSpace(cuerpo.Correo),
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, map[string]string{"estado": "ACTUALIZADO"})
}

func (manejador *ManejadorGobierno) CambiarEstadoDeUsuario(escritor http.ResponseWriter, peticion *http.Request) {
	var cuerpo struct {
		Estado string `json:"estado"`
	}
	if !web.DecodificarCuerpo(escritor, peticion, &cuerpo) {
		return
	}
	err := manejador.cambiarEstadoUsuario.Ejecutar(peticion.Context(), aplicacion.ComandoCambiarEstadoDeUsuario{
		IdentificadorUsuario: peticion.PathValue("id"),
		Estado:               strings.ToUpper(strings.TrimSpace(cuerpo.Estado)),
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, map[string]string{"estado": strings.ToUpper(cuerpo.Estado)})
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
		ColorPrimario        string `json:"color_primario"`
		ZonaHoraria          string `json:"zona_horaria"`
		Moneda               string `json:"moneda"`
		IdentificacionFiscal string `json:"identificacion_fiscal"`
		CorreoContacto       string `json:"correo_contacto"`
		Telefono             string `json:"telefono"`
	}
	if !web.DecodificarCuerpo(escritor, peticion, &cuerpo) {
		return
	}
	err := manejador.configurarEmpresa.Ejecutar(peticion.Context(), aplicacion.ComandoConfigurarEmpresa{
		IdentificadorEmpresa: empresaDe(peticion),
		ColorPrimario:        strings.TrimSpace(cuerpo.ColorPrimario),
		ZonaHoraria:          cuerpo.ZonaHoraria,
		Moneda:               strings.ToUpper(strings.TrimSpace(cuerpo.Moneda)),
		IdentificacionFiscal: strings.TrimSpace(cuerpo.IdentificacionFiscal),
		CorreoContacto:       strings.TrimSpace(cuerpo.CorreoContacto),
		Telefono:             strings.TrimSpace(cuerpo.Telefono),
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, map[string]string{"estado": "CONFIGURADA"})
}

func (manejador *ManejadorGobierno) SubirLogo(escritor http.ResponseWriter, peticion *http.Request) {
	peticion.Body = http.MaxBytesReader(escritor, peticion.Body, tamanoMaximoDeLogo)
	if err := peticion.ParseMultipartForm(tamanoMaximoDeLogo); err != nil {
		web.ResponderError(escritor, http.StatusBadRequest, "se esperaba un formulario multipart con el campo logo")
		return
	}
	archivo, _, err := peticion.FormFile("logo")
	if err != nil {
		web.ResponderError(escritor, http.StatusBadRequest, "falta el archivo en el campo logo")
		return
	}
	defer archivo.Close()
	datos, err := io.ReadAll(archivo)
	if err != nil {
		web.ResponderError(escritor, http.StatusBadRequest, "no se pudo leer el archivo")
		return
	}
	ruta, err := manejador.definirLogo.Ejecutar(peticion.Context(), aplicacion.ComandoDefinirLogo{
		IdentificadorEmpresa: empresaDe(peticion),
		Datos:                datos,
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, map[string]string{"LogoUrl": ruta})
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

func (manejador *ManejadorGobierno) MinasDeSesion(escritor http.ResponseWriter, peticion *http.Request) {
	sesion, _ := web.SesionDe(peticion.Context())
	acceso, err := manejador.minasDeSesion.Ejecutar(peticion.Context(), sesion.IdentificadorUsuario)
	if err != nil {
		web.ResponderError(escritor, http.StatusInternalServerError, err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, acceso)
}

func responderErrorDeSesion(escritor http.ResponseWriter, err error) {
	var bloqueado *aplicacion.ErrDemasiadosIntentos
	if errors.As(err, &bloqueado) {
		web.ResponderJson(escritor, http.StatusTooManyRequests, map[string]any{
			"error":                  "demasiados intentos, espera un momento",
			"reintentar_en_segundos": bloqueado.EsperaSegundos,
		})
		return
	}
	if errors.Is(err, aplicacion.ErrCredencialesInvalidas) {
		web.ResponderError(escritor, http.StatusUnauthorized, err.Error())
		return
	}
	log.Printf("error interno al iniciar sesion: %v", err)
	web.ResponderError(escritor, http.StatusServiceUnavailable, "no pudimos conectar con el servidor. Intenta de nuevo en un momento.")
}

func direccionRemota(peticion *http.Request) string {
	if reenviada := peticion.Header.Get("X-Forwarded-For"); reenviada != "" {
		return strings.TrimSpace(strings.Split(reenviada, ",")[0])
	}
	anfitrion, _, err := net.SplitHostPort(peticion.RemoteAddr)
	if err != nil {
		return peticion.RemoteAddr
	}
	return anfitrion
}

func empresaDe(peticion *http.Request) string {
	tenant, _ := contexto.TenantDe(peticion.Context())
	return tenant.Empresa.Texto()
}
