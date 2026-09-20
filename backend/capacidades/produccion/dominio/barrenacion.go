package dominio

import (
	"strings"

	"minas/compartido/identificador"
)

type Avance struct {
	id          identificador.Identificador
	actividad   string
	lugar       string
	noSecciones *int
	longitud    float64
	noBarrenos  int
	ejecutados  []identificador.Identificador
}

func (avance Avance) Identificador() identificador.Identificador { return avance.id }
func (avance Avance) Actividad() string                          { return avance.actividad }
func (avance Avance) Lugar() string                              { return avance.lugar }
func (avance Avance) NumeroDeSecciones() *int                    { return avance.noSecciones }
func (avance Avance) Longitud() float64                          { return avance.longitud }
func (avance Avance) NumeroDeBarrenos() int                      { return avance.noBarrenos }
func (avance Avance) Ejecutados() []identificador.Identificador  { return avance.ejecutados }

type Horometros struct {
	DieselInicial    *float64
	DieselFinal      *float64
	ElectricoInicial *float64
	ElectricoFinal   *float64
	PercusionInicial *float64
	PercusionFinal   *float64
}

type ParteDeBarrenacion struct {
	Cabecera
	tipoBarrenacion string
	idCapitanMina   identificador.Identificador
	idOperador      identificador.Identificador
	idSupervisor    *identificador.Identificador
	idAyudante      *identificador.Identificador
	idActividad     *identificador.Identificador
	horometros      Horometros
	avances         []Avance
}

func RegistrarParteDeBarrenacion(tipoBarrenacion string, idEmpresa, idMina, idObra, idEquipo,
	idCapitanMina, idOperador identificador.Identificador,
	idSupervisor, idAyudante, idActividad *identificador.Identificador,
	fecha, turno, observaciones string, horometros Horometros) (ParteDeBarrenacion, error) {
	if !EsValido(tipoBarrenacion, TiposDeBarrenacionValidos) {
		return ParteDeBarrenacion{}, ErrTipoDeBarrenacionInvalido
	}
	cabecera, err := nuevaCabecera(idEmpresa, idMina, idObra, idEquipo, fecha, turno, observaciones)
	if err != nil {
		return ParteDeBarrenacion{}, err
	}
	for _, par := range [][2]*float64{
		{horometros.DieselInicial, horometros.DieselFinal},
		{horometros.ElectricoInicial, horometros.ElectricoFinal},
		{horometros.PercusionInicial, horometros.PercusionFinal},
	} {
		if par[0] == nil || par[1] == nil {
			continue
		}
		if err := validarHorometros(*par[0], *par[1]); err != nil {
			return ParteDeBarrenacion{}, err
		}
	}
	return ParteDeBarrenacion{
		Cabecera:        cabecera,
		tipoBarrenacion: tipoBarrenacion,
		idCapitanMina:   idCapitanMina,
		idOperador:      idOperador,
		idSupervisor:    idSupervisor,
		idAyudante:      idAyudante,
		idActividad:     idActividad,
		horometros:      horometros,
	}, nil
}

func (parte *ParteDeBarrenacion) AgregarAvance(actividad, lugar string, noSecciones *int,
	longitud float64, noBarrenos int, ejecutados []identificador.Identificador) error {
	if !EsValido(actividad, ActividadesDeAvanceValidas) {
		return ErrActividadDeAvanceInvalida
	}
	sitio := strings.TrimSpace(lugar)
	if sitio == "" {
		return ErrLugarObligatorio
	}
	if longitud < 0 {
		return ErrLongitudInvalida
	}
	if noBarrenos < 0 {
		return ErrBarrenosInvalidos
	}
	parte.avances = append(parte.avances, Avance{
		id:          identificador.Nuevo(),
		actividad:   actividad,
		lugar:       sitio,
		noSecciones: noSecciones,
		longitud:    longitud,
		noBarrenos:  noBarrenos,
		ejecutados:  ejecutados,
	})
	return nil
}

func (parte ParteDeBarrenacion) TipoDeBarrenacion() string { return parte.tipoBarrenacion }
func (parte ParteDeBarrenacion) CapitanDeMina() identificador.Identificador {
	return parte.idCapitanMina
}
func (parte ParteDeBarrenacion) Operador() identificador.Identificador    { return parte.idOperador }
func (parte ParteDeBarrenacion) Supervisor() *identificador.Identificador { return parte.idSupervisor }
func (parte ParteDeBarrenacion) Ayudante() *identificador.Identificador   { return parte.idAyudante }
func (parte ParteDeBarrenacion) Actividad() *identificador.Identificador  { return parte.idActividad }
func (parte ParteDeBarrenacion) HorometrosDelTurno() Horometros           { return parte.horometros }
func (parte ParteDeBarrenacion) Avances() []Avance                        { return parte.avances }

func (parte ParteDeBarrenacion) MetrosBarrenados() float64 {
	var total float64
	for _, avance := range parte.avances {
		total += avance.longitud
	}
	return total
}
