package entrada

import (
	"log"
	"net/http"

	"minas/capacidades/produccion/aplicacion"
	"minas/capacidades/produccion/dominio"
	"minas/capacidades/produccion/puertos"
	"minas/compartido/paginacion"
	"minas/plataforma/contexto"
	"minas/plataforma/entrada/web"
)

type ManejadorProduccion struct {
	registrarCarga       *aplicacion.RegistrarParteDeCarga
	registrarBarrenacion *aplicacion.RegistrarParteDeBarrenacion
	registrarDemora      *aplicacion.RegistrarDemora
	consultas            *aplicacion.ConsultasDeProduccion
}

func NuevoManejadorProduccion(
	registrarCarga *aplicacion.RegistrarParteDeCarga,
	registrarBarrenacion *aplicacion.RegistrarParteDeBarrenacion,
	registrarDemora *aplicacion.RegistrarDemora,
	consultas *aplicacion.ConsultasDeProduccion,
) *ManejadorProduccion {
	return &ManejadorProduccion{
		registrarCarga:       registrarCarga,
		registrarBarrenacion: registrarBarrenacion,
		registrarDemora:      registrarDemora,
		consultas:            consultas,
	}
}

type cuerpoDeViaje struct {
	Desde         string   `json:"desde"`
	Hasta         string   `json:"hasta"`
	TipoDeMineral string   `json:"id_tipo_mineral"`
	Toneladas     *float64 `json:"toneladas"`
}

type cuerpoDeParteDeCarga struct {
	IdentificadorMina       string          `json:"id_mina"`
	IdentificadorObra       string          `json:"id_obra"`
	IdentificadorEquipo     string          `json:"id_equipo"`
	IdentificadorOperador   string          `json:"id_operador"`
	IdentificadorSupervisor string          `json:"id_supervisor"`
	Fecha                   string          `json:"fecha"`
	Turno                   string          `json:"turno"`
	Observaciones           string          `json:"observaciones"`
	HorometroInicial        float64         `json:"horometro_inicial"`
	HorometroFinal          float64         `json:"horometro_final"`
	Viajes                  []cuerpoDeViaje `json:"viajes"`
}

func (manejador *ManejadorProduccion) registrarCargaDe(modalidad string) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		var cuerpo cuerpoDeParteDeCarga
		if !web.DecodificarCuerpo(escritor, peticion, &cuerpo) {
			return
		}
		viajes := make([]aplicacion.ViajeCapturado, 0, len(cuerpo.Viajes))
		for _, viaje := range cuerpo.Viajes {
			viajes = append(viajes, aplicacion.ViajeCapturado{
				Desde:                    viaje.Desde,
				Hasta:                    viaje.Hasta,
				IdentificadorTipoMineral: viaje.TipoDeMineral,
				Toneladas:                viaje.Toneladas,
			})
		}
		identificadorParte, err := manejador.registrarCarga.Ejecutar(peticion.Context(), aplicacion.ComandoRegistrarParteDeCarga{
			Modalidad:               modalidad,
			IdentificadorEmpresa:    empresaDe(peticion),
			IdentificadorMina:       cuerpo.IdentificadorMina,
			IdentificadorObra:       cuerpo.IdentificadorObra,
			IdentificadorEquipo:     cuerpo.IdentificadorEquipo,
			IdentificadorOperador:   cuerpo.IdentificadorOperador,
			IdentificadorSupervisor: cuerpo.IdentificadorSupervisor,
			Fecha:                   cuerpo.Fecha,
			Turno:                   cuerpo.Turno,
			Observaciones:           cuerpo.Observaciones,
			HorometroInicial:        cuerpo.HorometroInicial,
			HorometroFinal:          cuerpo.HorometroFinal,
			Viajes:                  viajes,
		})
		if err != nil {
			web.ResponderError(escritor, codigoHttp(err), err.Error())
			return
		}
		web.ResponderJson(escritor, http.StatusCreated, map[string]string{"id": identificadorParte})
	}
}

func (manejador *ManejadorProduccion) RegistrarParteDeAcarreo(escritor http.ResponseWriter, peticion *http.Request) {
	manejador.registrarCargaDe(dominio.ModalidadAcarreo)(escritor, peticion)
}

func (manejador *ManejadorProduccion) RegistrarParteDeRezagado(escritor http.ResponseWriter, peticion *http.Request) {
	manejador.registrarCargaDe(dominio.ModalidadRezagado)(escritor, peticion)
}

func (manejador *ManejadorProduccion) listarCargaDe(modalidad string) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		partes, cursor, err := manejador.consultas.ListarPartesDeCarga(peticion.Context(), modalidad, filtroDe(peticion))
		if err != nil {
			responderErrorInterno(escritor, err)
			return
		}
		web.ResponderJson(escritor, http.StatusOK, map[string]any{"Elementos": partes, "SiguienteCursor": cursor})
	}
}

func (manejador *ManejadorProduccion) ListarPartesDeAcarreo(escritor http.ResponseWriter, peticion *http.Request) {
	manejador.listarCargaDe(dominio.ModalidadAcarreo)(escritor, peticion)
}

func (manejador *ManejadorProduccion) ListarPartesDeRezagado(escritor http.ResponseWriter, peticion *http.Request) {
	manejador.listarCargaDe(dominio.ModalidadRezagado)(escritor, peticion)
}

func (manejador *ManejadorProduccion) detalleDeCargaDe(modalidad string) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		detalle, encontrado, err := manejador.consultas.DetalleDeParteDeCarga(peticion.Context(), modalidad, peticion.PathValue("id"))
		if err != nil {
			responderErrorInterno(escritor, err)
			return
		}
		if !encontrado {
			web.ResponderError(escritor, http.StatusNotFound, "el parte no existe")
			return
		}
		web.ResponderJson(escritor, http.StatusOK, detalle)
	}
}

func (manejador *ManejadorProduccion) DetalleDeParteDeAcarreo(escritor http.ResponseWriter, peticion *http.Request) {
	manejador.detalleDeCargaDe(dominio.ModalidadAcarreo)(escritor, peticion)
}

func (manejador *ManejadorProduccion) DetalleDeParteDeRezagado(escritor http.ResponseWriter, peticion *http.Request) {
	manejador.detalleDeCargaDe(dominio.ModalidadRezagado)(escritor, peticion)
}

type cuerpoDeAvance struct {
	Actividad         string   `json:"actividad"`
	Lugar             string   `json:"lugar"`
	NumeroDeSecciones *int     `json:"no_secciones"`
	Longitud          float64  `json:"longitud"`
	NumeroDeBarrenos  int      `json:"no_barrenos"`
	Ejecutados        []string `json:"ejecutados"`
}

func (manejador *ManejadorProduccion) RegistrarParteDeBarrenacion(escritor http.ResponseWriter, peticion *http.Request) {
	var cuerpo struct {
		TipoDeBarrenacion        string           `json:"tipo_barrenacion"`
		IdentificadorMina        string           `json:"id_mina"`
		IdentificadorObra        string           `json:"id_obra"`
		IdentificadorEquipo      string           `json:"id_equipo"`
		IdentificadorCapitanMina string           `json:"id_capitan_mina"`
		IdentificadorOperador    string           `json:"id_operador"`
		IdentificadorSupervisor  string           `json:"id_supervisor"`
		IdentificadorAyudante    string           `json:"id_ayudante"`
		IdentificadorActividad   string           `json:"id_actividad"`
		Fecha                    string           `json:"fecha"`
		Turno                    string           `json:"turno"`
		Observaciones            string           `json:"observaciones"`
		DieselInicial            *float64         `json:"horometro_diesel_inicial"`
		DieselFinal              *float64         `json:"horometro_diesel_final"`
		ElectricoInicial         *float64         `json:"horometro_electrico_inicial"`
		ElectricoFinal           *float64         `json:"horometro_electrico_final"`
		PercusionInicial         *float64         `json:"horometro_percusion_inicial"`
		PercusionFinal           *float64         `json:"horometro_percusion_final"`
		Avances                  []cuerpoDeAvance `json:"avances"`
	}
	if !web.DecodificarCuerpo(escritor, peticion, &cuerpo) {
		return
	}
	avances := make([]aplicacion.AvanceCapturado, 0, len(cuerpo.Avances))
	for _, avance := range cuerpo.Avances {
		avances = append(avances, aplicacion.AvanceCapturado{
			Actividad:                avance.Actividad,
			Lugar:                    avance.Lugar,
			NumeroDeSecciones:        avance.NumeroDeSecciones,
			Longitud:                 avance.Longitud,
			NumeroDeBarrenos:         avance.NumeroDeBarrenos,
			IdentificadoresDeBarreno: avance.Ejecutados,
		})
	}
	identificadorParte, err := manejador.registrarBarrenacion.Ejecutar(peticion.Context(), aplicacion.ComandoRegistrarParteDeBarrenacion{
		TipoDeBarrenacion:        cuerpo.TipoDeBarrenacion,
		IdentificadorEmpresa:     empresaDe(peticion),
		IdentificadorMina:        cuerpo.IdentificadorMina,
		IdentificadorObra:        cuerpo.IdentificadorObra,
		IdentificadorEquipo:      cuerpo.IdentificadorEquipo,
		IdentificadorCapitanMina: cuerpo.IdentificadorCapitanMina,
		IdentificadorOperador:    cuerpo.IdentificadorOperador,
		IdentificadorSupervisor:  cuerpo.IdentificadorSupervisor,
		IdentificadorAyudante:    cuerpo.IdentificadorAyudante,
		IdentificadorActividad:   cuerpo.IdentificadorActividad,
		Fecha:                    cuerpo.Fecha,
		Turno:                    cuerpo.Turno,
		Observaciones:            cuerpo.Observaciones,
		Horometros: dominio.Horometros{
			DieselInicial:    cuerpo.DieselInicial,
			DieselFinal:      cuerpo.DieselFinal,
			ElectricoInicial: cuerpo.ElectricoInicial,
			ElectricoFinal:   cuerpo.ElectricoFinal,
			PercusionInicial: cuerpo.PercusionInicial,
			PercusionFinal:   cuerpo.PercusionFinal,
		},
		Avances: avances,
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusCreated, map[string]string{"id": identificadorParte})
}

func (manejador *ManejadorProduccion) ListarPartesDeBarrenacion(escritor http.ResponseWriter, peticion *http.Request) {
	partes, cursor, err := manejador.consultas.ListarPartesDeBarrenacion(peticion.Context(), filtroDe(peticion))
	if err != nil {
		responderErrorInterno(escritor, err)
		return
	}
	web.ResponderJson(escritor, http.StatusOK, map[string]any{"Elementos": partes, "SiguienteCursor": cursor})
}

func (manejador *ManejadorProduccion) DetalleDeParteDeBarrenacion(escritor http.ResponseWriter, peticion *http.Request) {
	detalle, encontrado, err := manejador.consultas.DetalleDeParteDeBarrenacion(peticion.Context(), peticion.PathValue("id"))
	if err != nil {
		responderErrorInterno(escritor, err)
		return
	}
	if !encontrado {
		web.ResponderError(escritor, http.StatusNotFound, "el parte no existe")
		return
	}
	web.ResponderJson(escritor, http.StatusOK, detalle)
}

func (manejador *ManejadorProduccion) RegistrarDemora(escritor http.ResponseWriter, peticion *http.Request) {
	var cuerpo struct {
		IdentificadorMina       string `json:"id_mina"`
		IdentificadorEquipo     string `json:"id_equipo"`
		IdentificadorTipoDemora string `json:"id_tipo_demora"`
		Fecha                   string `json:"fecha"`
		Turno                   string `json:"turno"`
		Minutos                 int    `json:"minutos"`
		Observacion             string `json:"observacion"`
	}
	if !web.DecodificarCuerpo(escritor, peticion, &cuerpo) {
		return
	}
	identificadorDemora, err := manejador.registrarDemora.Ejecutar(peticion.Context(), aplicacion.ComandoRegistrarDemora{
		IdentificadorEmpresa:    empresaDe(peticion),
		IdentificadorMina:       cuerpo.IdentificadorMina,
		IdentificadorEquipo:     cuerpo.IdentificadorEquipo,
		IdentificadorTipoDemora: cuerpo.IdentificadorTipoDemora,
		Fecha:                   cuerpo.Fecha,
		Turno:                   cuerpo.Turno,
		Minutos:                 cuerpo.Minutos,
		Observacion:             cuerpo.Observacion,
	})
	if err != nil {
		web.ResponderError(escritor, codigoHttp(err), err.Error())
		return
	}
	web.ResponderJson(escritor, http.StatusCreated, map[string]string{"id": identificadorDemora})
}

func (manejador *ManejadorProduccion) ListarDemoras(escritor http.ResponseWriter, peticion *http.Request) {
	demoras, cursor, err := manejador.consultas.ListarDemoras(peticion.Context(), filtroDe(peticion))
	if err != nil {
		responderErrorInterno(escritor, err)
		return
	}
	web.ResponderJson(escritor, http.StatusOK, map[string]any{"Elementos": demoras, "SiguienteCursor": cursor})
}

func (manejador *ManejadorProduccion) Indicadores(escritor http.ResponseWriter, peticion *http.Request) {
	indicadores, err := manejador.consultas.Indicadores(peticion.Context(), filtroDe(peticion))
	if err != nil {
		responderErrorInterno(escritor, err)
		return
	}
	web.ResponderJson(escritor, http.StatusOK, indicadores)
}

func filtroDe(peticion *http.Request) puertos.FiltroDeProduccion {
	consulta := peticion.URL.Query()
	filtro := puertos.FiltroDeProduccion{
		Mina:   consulta.Get("mina"),
		Obra:   consulta.Get("obra"),
		Equipo: consulta.Get("equipo"),
		Turno:  consulta.Get("turno"),
		Desde:  consulta.Get("desde"),
		Hasta:  consulta.Get("hasta"),
		Cursor: consulta.Get("cursor"),
		Limite: paginacion.LimiteSeguro(consulta.Get("limite")),
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
