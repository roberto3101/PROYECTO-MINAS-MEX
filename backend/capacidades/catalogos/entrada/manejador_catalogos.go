package entrada

import (
	"net/http"

	"minas/capacidades/catalogos/aplicacion"
	"minas/plataforma/contexto"
	"minas/plataforma/entrada/web"
)

type ManejadorCatalogos struct {
	crearMina         *aplicacion.CrearMina
	listarMinas       *aplicacion.ListarMinas
	contratarEmpleado *aplicacion.ContratarEmpleado
	listarEmpleados   *aplicacion.ListarEmpleados
	darDeAltaEquipo   *aplicacion.DarDeAltaEquipo
	listarEquipos     *aplicacion.ListarEquipos
	listarTipos       *aplicacion.ListarTiposDeEquipo
	listarModulos     *aplicacion.ListarModulosDeTrabajo
}

func NuevoManejadorCatalogos(
	crearMina *aplicacion.CrearMina,
	listarMinas *aplicacion.ListarMinas,
	contratarEmpleado *aplicacion.ContratarEmpleado,
	listarEmpleados *aplicacion.ListarEmpleados,
	darDeAltaEquipo *aplicacion.DarDeAltaEquipo,
	listarEquipos *aplicacion.ListarEquipos,
	listarTipos *aplicacion.ListarTiposDeEquipo,
	listarModulos *aplicacion.ListarModulosDeTrabajo,
) *ManejadorCatalogos {
	return &ManejadorCatalogos{
		crearMina:         crearMina,
		listarMinas:       listarMinas,
		contratarEmpleado: contratarEmpleado,
		listarEmpleados:   listarEmpleados,
		darDeAltaEquipo:   darDeAltaEquipo,
		listarEquipos:     listarEquipos,
		listarTipos:       listarTipos,
		listarModulos:     listarModulos,
	}
}

func (manejador *ManejadorCatalogos) CrearMina(escritor http.ResponseWriter, peticion *http.Request) {
	var cuerpo struct {
		Nombre string `json:"nombre"`
		Area   string `json:"area"`
	}
	if !web.DecodificarCuerpo(escritor, peticion, &cuerpo) {
		return
	}
	identificadorMina, err := manejador.crearMina.Ejecutar(peticion.Context(), aplicacion.ComandoCrearMina{
		IdentificadorEmpresa: empresaDe(peticion),
		Nombre:               cuerpo.Nombre,
		Area:                 cuerpo.Area,
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusCreated, map[string]string{"id": identificadorMina})
}

func (manejador *ManejadorCatalogos) ListarMinas(escritor http.ResponseWriter, peticion *http.Request) {
	minas, err := manejador.listarMinas.Ejecutar(peticion.Context())
	if err != nil {
		web.ResponderError(escritor, http.StatusInternalServerError, err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, minas)
}

func (manejador *ManejadorCatalogos) ContratarEmpleado(escritor http.ResponseWriter, peticion *http.Request) {
	var cuerpo struct {
		IdentificadorMina string `json:"id_mina"`
		NumeroNomina      string `json:"numero_nomina"`
		NombreCompleto    string `json:"nombre_completo"`
	}
	if !web.DecodificarCuerpo(escritor, peticion, &cuerpo) {
		return
	}
	identificadorEmpleado, err := manejador.contratarEmpleado.Ejecutar(peticion.Context(), aplicacion.ComandoContratarEmpleado{
		IdentificadorEmpresa: empresaDe(peticion),
		IdentificadorMina:    cuerpo.IdentificadorMina,
		NumeroNomina:         cuerpo.NumeroNomina,
		NombreCompleto:       cuerpo.NombreCompleto,
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusCreated, map[string]string{"id": identificadorEmpleado})
}

func (manejador *ManejadorCatalogos) ListarEmpleados(escritor http.ResponseWriter, peticion *http.Request) {
	empleados, err := manejador.listarEmpleados.Ejecutar(peticion.Context())
	if err != nil {
		web.ResponderError(escritor, http.StatusInternalServerError, err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, empleados)
}

func (manejador *ManejadorCatalogos) DarDeAltaEquipo(escritor http.ResponseWriter, peticion *http.Request) {
	var cuerpo struct {
		IdentificadorMina          string `json:"id_mina"`
		IdentificadorTipoEquipo    string `json:"id_tipo_equipo"`
		IdentificadorModuloTrabajo string `json:"id_modulo_trabajo"`
		Codigo                     string `json:"codigo"`
		Fabricante                 string `json:"fabricante"`
	}
	if !web.DecodificarCuerpo(escritor, peticion, &cuerpo) {
		return
	}
	identificadorEquipo, err := manejador.darDeAltaEquipo.Ejecutar(peticion.Context(), aplicacion.ComandoDarDeAltaEquipo{
		IdentificadorEmpresa:       empresaDe(peticion),
		IdentificadorMina:          cuerpo.IdentificadorMina,
		IdentificadorTipoEquipo:    cuerpo.IdentificadorTipoEquipo,
		IdentificadorModuloTrabajo: cuerpo.IdentificadorModuloTrabajo,
		Codigo:                     cuerpo.Codigo,
		Fabricante:                 cuerpo.Fabricante,
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusCreated, map[string]string{"id": identificadorEquipo})
}

func (manejador *ManejadorCatalogos) ListarEquipos(escritor http.ResponseWriter, peticion *http.Request) {
	equipos, err := manejador.listarEquipos.Ejecutar(peticion.Context())
	if err != nil {
		web.ResponderError(escritor, http.StatusInternalServerError, err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, equipos)
}

func (manejador *ManejadorCatalogos) ListarTiposDeEquipo(escritor http.ResponseWriter, peticion *http.Request) {
	tipos, err := manejador.listarTipos.Ejecutar(peticion.Context())
	if err != nil {
		web.ResponderError(escritor, http.StatusInternalServerError, err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, tipos)
}

func (manejador *ManejadorCatalogos) ListarModulosDeTrabajo(escritor http.ResponseWriter, peticion *http.Request) {
	modulos, err := manejador.listarModulos.Ejecutar(peticion.Context())
	if err != nil {
		web.ResponderError(escritor, http.StatusInternalServerError, err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, modulos)
}

func empresaDe(peticion *http.Request) string {
	tenant, _ := contexto.TenantDe(peticion.Context())
	return tenant.Empresa.Texto()
}
