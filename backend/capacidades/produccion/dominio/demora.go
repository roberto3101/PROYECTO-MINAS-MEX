package dominio

import (
	"strings"
	"time"

	"minas/compartido/identificador"
)

type Demora struct {
	id           identificador.Identificador
	idEmpresa    identificador.Identificador
	idMina       identificador.Identificador
	idEquipo     identificador.Identificador
	idTipoDemora identificador.Identificador
	fecha        time.Time
	turno        string
	minutos      int
	observacion  string
}

func RegistrarDemora(idEmpresa, idMina, idEquipo, idTipoDemora identificador.Identificador,
	fecha, turno string, minutos int, observacion string) (Demora, error) {
	if !EsValido(turno, TurnosValidos) {
		return Demora{}, ErrTurnoNoReconocido
	}
	dia, err := FechaDeCaptura(fecha)
	if err != nil {
		return Demora{}, err
	}
	if minutos < 0 || minutos > minutosMaximosPorTurno {
		return Demora{}, ErrMinutosInvalidos
	}
	return Demora{
		id:           identificador.Nuevo(),
		idEmpresa:    idEmpresa,
		idMina:       idMina,
		idEquipo:     idEquipo,
		idTipoDemora: idTipoDemora,
		fecha:        dia,
		turno:        turno,
		minutos:      minutos,
		observacion:  strings.TrimSpace(observacion),
	}, nil
}

func (demora Demora) Identificador() identificador.Identificador { return demora.id }
func (demora Demora) Empresa() identificador.Identificador       { return demora.idEmpresa }
func (demora Demora) Mina() identificador.Identificador          { return demora.idMina }
func (demora Demora) Equipo() identificador.Identificador        { return demora.idEquipo }
func (demora Demora) TipoDeDemora() identificador.Identificador  { return demora.idTipoDemora }
func (demora Demora) Fecha() time.Time                           { return demora.fecha }
func (demora Demora) Turno() string                              { return demora.turno }
func (demora Demora) Minutos() int                               { return demora.minutos }
func (demora Demora) Observacion() string                        { return demora.observacion }
