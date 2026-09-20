package entrada

import (
	"log"
	"net/http"

	"minas/capacidades/seguridad/aplicacion"
	"minas/capacidades/seguridad/puertos"
	"minas/compartido/paginacion"
	"minas/plataforma/contexto"
	"minas/plataforma/entrada/web"
)

type ManejadorSeguridad struct {
	reportarIncidente *aplicacion.ReportarIncidente
	consultas         *aplicacion.ConsultasDeSeguridad
}

func NuevoManejadorSeguridad(reportarIncidente *aplicacion.ReportarIncidente, consultas *aplicacion.ConsultasDeSeguridad) *ManejadorSeguridad {
	return &ManejadorSeguridad{reportarIncidente: reportarIncidente, consultas: consultas}
}

func (manejador *ManejadorSeguridad) ReportarIncidente(escritor http.ResponseWriter, peticion *http.Request) {
	var cuerpo struct {
		IdentificadorMina          string `json:"id_mina"`
		IdentificadorObra          string `json:"id_obra"`
		IdentificadorTipoIncidente string `json:"id_tipo_incidente"`
		IdentificadorReportadoPor  string `json:"id_reportado_por"`
		Fecha                      string `json:"fecha"`
		Turno                      string `json:"turno"`
		Severidad                  string `json:"severidad"`
		Descripcion                string `json:"descripcion"`
		AccionInmediata            string `json:"accion_inmediata"`
	}
	if !web.DecodificarCuerpo(escritor, peticion, &cuerpo) {
		return
	}
	identificadorIncidente, err := manejador.reportarIncidente.Ejecutar(peticion.Context(), aplicacion.ComandoReportarIncidente{
		IdentificadorEmpresa:       empresaDe(peticion),
		IdentificadorMina:          cuerpo.IdentificadorMina,
		IdentificadorObra:          cuerpo.IdentificadorObra,
		IdentificadorTipoIncidente: cuerpo.IdentificadorTipoIncidente,
		IdentificadorReportadoPor:  cuerpo.IdentificadorReportadoPor,
		Fecha:                      cuerpo.Fecha,
		Turno:                      cuerpo.Turno,
		Severidad:                  cuerpo.Severidad,
		Descripcion:                cuerpo.Descripcion,
		AccionInmediata:            cuerpo.AccionInmediata,
	})
	if err != nil {
		web.ResponderError(escritor, http.StatusBadRequest, err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusCreated, map[string]string{"id": identificadorIncidente})
}

func (manejador *ManejadorSeguridad) ListarIncidentes(escritor http.ResponseWriter, peticion *http.Request) {
	incidentes, cursor, err := manejador.consultas.ListarIncidentes(peticion.Context(), filtroDe(peticion))
	if err != nil {
		responderErrorInterno(escritor, err)
		return
	}
	web.ResponderJson(escritor, http.StatusOK, map[string]any{"Elementos": incidentes, "SiguienteCursor": cursor})
}

func (manejador *ManejadorSeguridad) ListarTiposDeIncidente(escritor http.ResponseWriter, peticion *http.Request) {
	tipos, err := manejador.consultas.ListarTiposDeIncidente(peticion.Context())
	if err != nil {
		responderErrorInterno(escritor, err)
		return
	}
	web.ResponderJson(escritor, http.StatusOK, tipos)
}

func (manejador *ManejadorSeguridad) Indicadores(escritor http.ResponseWriter, peticion *http.Request) {
	conteo, err := manejador.consultas.ConteoPorSeveridad(peticion.Context(), filtroDe(peticion))
	if err != nil {
		responderErrorInterno(escritor, err)
		return
	}
	web.ResponderJson(escritor, http.StatusOK, conteo)
}

func filtroDe(peticion *http.Request) puertos.FiltroDeIncidentes {
	consulta := peticion.URL.Query()
	filtro := puertos.FiltroDeIncidentes{
		Mina:      consulta.Get("mina"),
		Severidad: consulta.Get("severidad"),
		Desde:     consulta.Get("desde"),
		Hasta:     consulta.Get("hasta"),
		Cursor:    consulta.Get("cursor"),
		Limite:    paginacion.LimiteSeguro(consulta.Get("limite")),
	}
	tenant, _ := contexto.TenantDe(peticion.Context())
	if !tenant.AlcanceGlobalDeMinas {
		filtro.RestringirPorMina = true
		filtro.MinasPermitidas = tenant.MinasPermitidas
		if filtro.Mina != "" && !contiene(tenant.MinasPermitidas, filtro.Mina) {
			filtro.MinasPermitidas = []string{}
		}
	}
	return filtro
}

func contiene(valores []string, objetivo string) bool {
	for _, valor := range valores {
		if valor == objetivo {
			return true
		}
	}
	return false
}

func empresaDe(peticion *http.Request) string {
	tenant, _ := contexto.TenantDe(peticion.Context())
	return tenant.Empresa.Texto()
}

func responderErrorInterno(escritor http.ResponseWriter, err error) {
	log.Printf("error interno: %v", err)
	web.ResponderError(escritor, http.StatusInternalServerError, "no se pudo completar la operacion")
}
