package entrada

import (
	"net/http"

	"minas/capacidades/catalogos/aplicacion"
	"minas/capacidades/catalogos/puertos"
	"minas/compartido/paginacion"
	"minas/plataforma/contexto"
	"minas/plataforma/entrada/web"
)

type ManejadorCatalogos struct {
	crearMina             *aplicacion.CrearMina
	listarMinas           *aplicacion.ListarMinas
	detalleMina           *aplicacion.DetalleDeMina
	cambiarEstadoMina     *aplicacion.CambiarEstadoDeMina
	contratarEmpleado     *aplicacion.ContratarEmpleado
	listarEmpleados       *aplicacion.ListarEmpleados
	detalleEmpleado       *aplicacion.DetalleDeEmpleado
	cambiarEstadoEmpleado *aplicacion.CambiarEstadoDeEmpleado
	darDeAltaEquipo       *aplicacion.DarDeAltaEquipo
	listarEquipos         *aplicacion.ListarEquipos
	detalleEquipo         *aplicacion.DetalleDeEquipo
	cambiarEstadoEquipo   *aplicacion.CambiarEstadoDeEquipo
	listarTipos           *aplicacion.ListarTiposDeEquipo
	listarModulos         *aplicacion.ListarModulosDeTrabajo
	listarDepartamentos   *aplicacion.ListarDepartamentos
	listarPuestos         *aplicacion.ListarPuestos
	listarActividades     *aplicacion.ListarActividades
}

func NuevoManejadorCatalogos(
	crearMina *aplicacion.CrearMina,
	listarMinas *aplicacion.ListarMinas,
	detalleMina *aplicacion.DetalleDeMina,
	cambiarEstadoMina *aplicacion.CambiarEstadoDeMina,
	contratarEmpleado *aplicacion.ContratarEmpleado,
	listarEmpleados *aplicacion.ListarEmpleados,
	detalleEmpleado *aplicacion.DetalleDeEmpleado,
	cambiarEstadoEmpleado *aplicacion.CambiarEstadoDeEmpleado,
	darDeAltaEquipo *aplicacion.DarDeAltaEquipo,
	listarEquipos *aplicacion.ListarEquipos,
	detalleEquipo *aplicacion.DetalleDeEquipo,
	cambiarEstadoEquipo *aplicacion.CambiarEstadoDeEquipo,
	listarTipos *aplicacion.ListarTiposDeEquipo,
	listarModulos *aplicacion.ListarModulosDeTrabajo,
	listarDepartamentos *aplicacion.ListarDepartamentos,
	listarPuestos *aplicacion.ListarPuestos,
	listarActividades *aplicacion.ListarActividades,
) *ManejadorCatalogos {
	return &ManejadorCatalogos{
		crearMina:             crearMina,
		listarMinas:           listarMinas,
		detalleMina:           detalleMina,
		cambiarEstadoMina:     cambiarEstadoMina,
		contratarEmpleado:     contratarEmpleado,
		listarEmpleados:       listarEmpleados,
		detalleEmpleado:       detalleEmpleado,
		cambiarEstadoEmpleado: cambiarEstadoEmpleado,
		darDeAltaEquipo:       darDeAltaEquipo,
		listarEquipos:         listarEquipos,
		detalleEquipo:         detalleEquipo,
		cambiarEstadoEquipo:   cambiarEstadoEquipo,
		listarTipos:           listarTipos,
		listarModulos:         listarModulos,
		listarDepartamentos:   listarDepartamentos,
		listarPuestos:         listarPuestos,
		listarActividades:     listarActividades,
	}
}

func (manejador *ManejadorCatalogos) CrearMina(escritor http.ResponseWriter, peticion *http.Request) {
	var cuerpo struct {
		Nombre           string   `json:"nombre"`
		Area             string   `json:"area"`
		Niveles          string   `json:"niveles"`
		DensidadPromedio *float64 `json:"densidad_promedio"`
		SeccionObraLargo *float64 `json:"seccion_obra_largo"`
		SeccionObraAncho *float64 `json:"seccion_obra_ancho"`
		SeccionVetaLargo *float64 `json:"seccion_veta_largo"`
		SeccionVetaAncho *float64 `json:"seccion_veta_ancho"`
	}
	if !web.DecodificarCuerpo(escritor, peticion, &cuerpo) {
		return
	}
	identificadorMina, err := manejador.crearMina.Ejecutar(peticion.Context(), aplicacion.ComandoCrearMina{
		IdentificadorEmpresa: empresaDe(peticion),
		Nombre:               cuerpo.Nombre,
		Area:                 cuerpo.Area,
		Niveles:              cuerpo.Niveles,
		DensidadPromedio:     cuerpo.DensidadPromedio,
		SeccionObraLargo:     cuerpo.SeccionObraLargo,
		SeccionObraAncho:     cuerpo.SeccionObraAncho,
		SeccionVetaLargo:     cuerpo.SeccionVetaLargo,
		SeccionVetaAncho:     cuerpo.SeccionVetaAncho,
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusCreated, map[string]string{"id": identificadorMina})
}

func (manejador *ManejadorCatalogos) ListarMinas(escritor http.ResponseWriter, peticion *http.Request) {
	minas, cursor, err := manejador.listarMinas.Ejecutar(peticion.Context(), filtroDe(peticion))
	if err != nil {
		web.ResponderError(escritor, http.StatusInternalServerError, err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, map[string]any{"Elementos": minas, "SiguienteCursor": cursor})
}

func (manejador *ManejadorCatalogos) DetalleDeMina(escritor http.ResponseWriter, peticion *http.Request) {
	detalle, encontrada, err := manejador.detalleMina.Ejecutar(peticion.Context(), peticion.PathValue("id"))
	if err != nil {
		web.ResponderError(escritor, http.StatusInternalServerError, err.Error())
		return
	}
	if !encontrada {
		web.ResponderError(escritor, http.StatusNotFound, "no encontrado")
		return
	}
	web.ResponderJson(escritor, http.StatusOK, detalle)
}

func (manejador *ManejadorCatalogos) CambiarEstadoDeMina(escritor http.ResponseWriter, peticion *http.Request) {
	var cuerpo struct {
		Estado string `json:"estado"`
	}
	if !web.DecodificarCuerpo(escritor, peticion, &cuerpo) {
		return
	}
	err := manejador.cambiarEstadoMina.Ejecutar(peticion.Context(), aplicacion.ComandoCambiarEstadoDeMina{
		IdentificadorMina: peticion.PathValue("id"),
		Estado:            cuerpo.Estado,
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, map[string]string{"estado": cuerpo.Estado})
}

func (manejador *ManejadorCatalogos) ContratarEmpleado(escritor http.ResponseWriter, peticion *http.Request) {
	var cuerpo struct {
		IdentificadorMina         string `json:"id_mina"`
		NumeroNomina              string `json:"numero_nomina"`
		NombreCompleto            string `json:"nombre_completo"`
		IdentificadorDepartamento string `json:"id_departamento"`
		IdentificadorPuesto       string `json:"id_puesto"`
		IdentificadorActividad    string `json:"id_actividad"`
		CentroCosto               string `json:"centro_costo"`
		GerenteACargo             string `json:"gerente_a_cargo"`
		Grupo                     string `json:"grupo"`
	}
	if !web.DecodificarCuerpo(escritor, peticion, &cuerpo) {
		return
	}
	identificadorEmpleado, err := manejador.contratarEmpleado.Ejecutar(peticion.Context(), aplicacion.ComandoContratarEmpleado{
		IdentificadorEmpresa:      empresaDe(peticion),
		IdentificadorMina:         cuerpo.IdentificadorMina,
		NumeroNomina:              cuerpo.NumeroNomina,
		NombreCompleto:            cuerpo.NombreCompleto,
		IdentificadorDepartamento: cuerpo.IdentificadorDepartamento,
		IdentificadorPuesto:       cuerpo.IdentificadorPuesto,
		IdentificadorActividad:    cuerpo.IdentificadorActividad,
		CentroCosto:               cuerpo.CentroCosto,
		GerenteACargo:             cuerpo.GerenteACargo,
		Grupo:                     cuerpo.Grupo,
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusCreated, map[string]string{"id": identificadorEmpleado})
}

func (manejador *ManejadorCatalogos) ListarEmpleados(escritor http.ResponseWriter, peticion *http.Request) {
	empleados, cursor, err := manejador.listarEmpleados.Ejecutar(peticion.Context(), filtroDe(peticion))
	if err != nil {
		web.ResponderError(escritor, http.StatusInternalServerError, err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, map[string]any{"Elementos": empleados, "SiguienteCursor": cursor})
}

func (manejador *ManejadorCatalogos) DetalleDeEmpleado(escritor http.ResponseWriter, peticion *http.Request) {
	detalle, encontrado, err := manejador.detalleEmpleado.Ejecutar(peticion.Context(), peticion.PathValue("id"))
	if err != nil {
		web.ResponderError(escritor, http.StatusInternalServerError, err.Error())
		return
	}
	if !encontrado {
		web.ResponderError(escritor, http.StatusNotFound, "no encontrado")
		return
	}
	web.ResponderJson(escritor, http.StatusOK, detalle)
}

func (manejador *ManejadorCatalogos) CambiarEstadoDeEmpleado(escritor http.ResponseWriter, peticion *http.Request) {
	var cuerpo struct {
		Estado string `json:"estado"`
	}
	if !web.DecodificarCuerpo(escritor, peticion, &cuerpo) {
		return
	}
	err := manejador.cambiarEstadoEmpleado.Ejecutar(peticion.Context(), aplicacion.ComandoCambiarEstadoDeEmpleado{
		IdentificadorEmpleado: peticion.PathValue("id"),
		Estado:                cuerpo.Estado,
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, map[string]string{"estado": cuerpo.Estado})
}

func (manejador *ManejadorCatalogos) DarDeAltaEquipo(escritor http.ResponseWriter, peticion *http.Request) {
	var cuerpo struct {
		IdentificadorMina          string   `json:"id_mina"`
		IdentificadorTipoEquipo    string   `json:"id_tipo_equipo"`
		IdentificadorModuloTrabajo string   `json:"id_modulo_trabajo"`
		Codigo                     string   `json:"codigo"`
		Descripcion                string   `json:"descripcion"`
		Fabricante                 string   `json:"fabricante"`
		Modelo                     string   `json:"modelo"`
		NumeroSerie                string   `json:"numero_serie"`
		AnioFabricacion            *int     `json:"anio_fabricacion"`
		FechaIngresoMina           string   `json:"fecha_ingreso_mina"`
		ModeloPerforadora          string   `json:"modelo_perforadora"`
		CapacidadLongitud          *float64 `json:"capacidad_longitud"`
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
		Descripcion:                cuerpo.Descripcion,
		ModeloPerforadora:          cuerpo.ModeloPerforadora,
		CapacidadLongitud:          cuerpo.CapacidadLongitud,
		Fabricante:                 cuerpo.Fabricante,
		Modelo:                     cuerpo.Modelo,
		NumeroSerie:                cuerpo.NumeroSerie,
		AnioFabricacion:            cuerpo.AnioFabricacion,
		FechaIngresoMina:           cuerpo.FechaIngresoMina,
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusCreated, map[string]string{"id": identificadorEquipo})
}

func (manejador *ManejadorCatalogos) ListarEquipos(escritor http.ResponseWriter, peticion *http.Request) {
	equipos, cursor, err := manejador.listarEquipos.Ejecutar(peticion.Context(), filtroDe(peticion))
	if err != nil {
		web.ResponderError(escritor, http.StatusInternalServerError, err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, map[string]any{"Elementos": equipos, "SiguienteCursor": cursor})
}

func (manejador *ManejadorCatalogos) DetalleDeEquipo(escritor http.ResponseWriter, peticion *http.Request) {
	detalle, encontrado, err := manejador.detalleEquipo.Ejecutar(peticion.Context(), peticion.PathValue("id"))
	if err != nil {
		web.ResponderError(escritor, http.StatusInternalServerError, err.Error())
		return
	}
	if !encontrado {
		web.ResponderError(escritor, http.StatusNotFound, "no encontrado")
		return
	}
	web.ResponderJson(escritor, http.StatusOK, detalle)
}

func (manejador *ManejadorCatalogos) CambiarEstadoDeEquipo(escritor http.ResponseWriter, peticion *http.Request) {
	var cuerpo struct {
		Estado string `json:"estado"`
	}
	if !web.DecodificarCuerpo(escritor, peticion, &cuerpo) {
		return
	}
	err := manejador.cambiarEstadoEquipo.Ejecutar(peticion.Context(), aplicacion.ComandoCambiarEstadoDeEquipo{
		IdentificadorEquipo: peticion.PathValue("id"),
		Estado:              cuerpo.Estado,
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, map[string]string{"estado": cuerpo.Estado})
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

func (manejador *ManejadorCatalogos) ListarDepartamentos(escritor http.ResponseWriter, peticion *http.Request) {
	departamentos, err := manejador.listarDepartamentos.Ejecutar(peticion.Context())
	if err != nil {
		web.ResponderError(escritor, http.StatusInternalServerError, err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, departamentos)
}

func (manejador *ManejadorCatalogos) ListarPuestos(escritor http.ResponseWriter, peticion *http.Request) {
	puestos, err := manejador.listarPuestos.Ejecutar(peticion.Context())
	if err != nil {
		web.ResponderError(escritor, http.StatusInternalServerError, err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, puestos)
}

func (manejador *ManejadorCatalogos) ListarActividades(escritor http.ResponseWriter, peticion *http.Request) {
	actividades, err := manejador.listarActividades.Ejecutar(peticion.Context())
	if err != nil {
		web.ResponderError(escritor, http.StatusInternalServerError, err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusOK, actividades)
}

func filtroDe(peticion *http.Request) puertos.FiltroDeCatalogo {
	consulta := peticion.URL.Query()
	return puertos.FiltroDeCatalogo{
		Busqueda: consulta.Get("busqueda"),
		Estado:   consulta.Get("estado"),
		Cursor:   consulta.Get("cursor"),
		Limite:   paginacion.LimiteSeguro(consulta.Get("limite")),
	}
}

func empresaDe(peticion *http.Request) string {
	tenant, _ := contexto.TenantDe(peticion.Context())
	return tenant.Empresa.Texto()
}
