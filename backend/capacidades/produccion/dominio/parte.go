package dominio

import (
	"errors"
	"strings"
	"time"

	"minas/compartido/identificador"
)

const (
	TurnoMatutino   = "M"
	TurnoVespertino = "V"
	TurnoNocturno   = "N"
)

var TurnosValidos = []string{TurnoMatutino, TurnoVespertino, TurnoNocturno}

const (
	ModalidadAcarreo  = "ACARREO"
	ModalidadRezagado = "REZAGADO"
)

const (
	BarrenacionLineal    = "LINEAL"
	BarrenacionMaqPierna = "MAQ_PIERNA"
	BarrenacionAnclador  = "ANCLADOR"
	BarrenacionLarga     = "LARGA"
)

var TiposDeBarrenacionValidos = []string{
	BarrenacionLineal, BarrenacionMaqPierna, BarrenacionAnclador, BarrenacionLarga,
}

const (
	AvanceBarrenacion   = "BARRENACION"
	AvanceReBarrenacion = "RE_BARRENACION"
	AvanceEscareo       = "ESCAREO"
)

var ActividadesDeAvanceValidas = []string{AvanceBarrenacion, AvanceReBarrenacion, AvanceEscareo}

var (
	ErrTurnoNoReconocido          = errors.New("el turno debe ser M (matutino), V (vespertino) o N (nocturno)")
	ErrHorometroInconsistente     = errors.New("el horometro final no puede ser menor que el inicial")
	ErrHorometroNegativo          = errors.New("el horometro no puede ser negativo")
	ErrFechaEnElFuturo            = errors.New("no se puede capturar un parte con fecha futura")
	ErrFechaInvalida              = errors.New("la fecha debe tener el formato AAAA-MM-DD")
	ErrToneladasInvalidas         = errors.New("las toneladas deben ser mayores que cero")
	ErrOrigenYDestinoObligatorios = errors.New("cada viaje necesita origen y destino")
	ErrTipoDeBarrenacionInvalido  = errors.New("el tipo de barrenacion no es reconocido")
	ErrActividadDeAvanceInvalida  = errors.New("la actividad del avance no es reconocida")
	ErrLugarObligatorio           = errors.New("el lugar del avance es obligatorio")
	ErrLongitudInvalida           = errors.New("la longitud no puede ser negativa")
	ErrBarrenosInvalidos          = errors.New("el numero de barrenos no puede ser negativo")
	ErrMinutosInvalidos           = errors.New("los minutos de la demora no pueden ser negativos ni pasar de un turno")
	ErrParteNoEncontrado          = errors.New("el parte no existe")
	ErrModalidadNoReconocida      = errors.New("la modalidad de carga no es reconocida")
)

const minutosMaximosPorTurno = 12 * 60

func EsValido(valor string, validos []string) bool {
	for _, candidato := range validos {
		if candidato == valor {
			return true
		}
	}
	return false
}

func FechaDeCaptura(valor string) (time.Time, error) {
	fecha, err := time.Parse("2006-01-02", strings.TrimSpace(valor))
	if err != nil {
		return time.Time{}, ErrFechaInvalida
	}
	hoy := time.Now()
	if fecha.After(time.Date(hoy.Year(), hoy.Month(), hoy.Day(), 23, 59, 59, 0, fecha.Location())) {
		return time.Time{}, ErrFechaEnElFuturo
	}
	return fecha, nil
}

type Cabecera struct {
	id            identificador.Identificador
	idEmpresa     identificador.Identificador
	idMina        identificador.Identificador
	idObra        identificador.Identificador
	idEquipo      identificador.Identificador
	fecha         time.Time
	turno         string
	observaciones string
}

func (cabecera Cabecera) Identificador() identificador.Identificador { return cabecera.id }
func (cabecera Cabecera) Empresa() identificador.Identificador       { return cabecera.idEmpresa }
func (cabecera Cabecera) Mina() identificador.Identificador          { return cabecera.idMina }
func (cabecera Cabecera) Obra() identificador.Identificador          { return cabecera.idObra }
func (cabecera Cabecera) Equipo() identificador.Identificador        { return cabecera.idEquipo }
func (cabecera Cabecera) Fecha() time.Time                           { return cabecera.fecha }
func (cabecera Cabecera) Turno() string                              { return cabecera.turno }
func (cabecera Cabecera) Observaciones() string                      { return cabecera.observaciones }

func nuevaCabecera(idEmpresa, idMina, idObra, idEquipo identificador.Identificador,
	fecha, turno, observaciones string) (Cabecera, error) {
	if !EsValido(turno, TurnosValidos) {
		return Cabecera{}, ErrTurnoNoReconocido
	}
	dia, err := FechaDeCaptura(fecha)
	if err != nil {
		return Cabecera{}, err
	}
	return Cabecera{
		id:            identificador.Nuevo(),
		idEmpresa:     idEmpresa,
		idMina:        idMina,
		idObra:        idObra,
		idEquipo:      idEquipo,
		fecha:         dia,
		turno:         turno,
		observaciones: strings.TrimSpace(observaciones),
	}, nil
}

func validarHorometros(inicial, final float64) error {
	if inicial < 0 || final < 0 {
		return ErrHorometroNegativo
	}
	if final < inicial {
		return ErrHorometroInconsistente
	}
	return nil
}
