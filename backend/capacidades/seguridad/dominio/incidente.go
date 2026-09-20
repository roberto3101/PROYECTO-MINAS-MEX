package dominio

import (
	"errors"
	"strings"
	"time"

	"minas/compartido/identificador"
)

const (
	SeveridadBaja    = "BAJA"
	SeveridadMedia   = "MEDIA"
	SeveridadAlta    = "ALTA"
	SeveridadCritica = "CRITICA"
)

var SeveridadesValidas = []string{SeveridadBaja, SeveridadMedia, SeveridadAlta, SeveridadCritica}

var TurnosValidos = []string{"M", "V", "N"}

var (
	ErrTurnoNoReconocido      = errors.New("el turno debe ser M (matutino), V (vespertino) o N (nocturno)")
	ErrSeveridadNoReconocida  = errors.New("la severidad debe ser BAJA, MEDIA, ALTA o CRITICA")
	ErrDescripcionObligatoria = errors.New("la descripcion del incidente es obligatoria")
	ErrFechaInvalida          = errors.New("la fecha debe tener el formato AAAA-MM-DD")
	ErrFechaEnElFuturo        = errors.New("no se puede reportar un incidente con fecha futura")
)

func EsValido(valor string, validos []string) bool {
	for _, candidato := range validos {
		if candidato == valor {
			return true
		}
	}
	return false
}

type Incidente struct {
	id              identificador.Identificador
	idEmpresa       identificador.Identificador
	idMina          identificador.Identificador
	idObra          *identificador.Identificador
	idTipoIncidente identificador.Identificador
	idReportadoPor  *identificador.Identificador
	fecha           time.Time
	turno           string
	severidad       string
	descripcion     string
	accionInmediata string
}

func ReportarIncidente(idEmpresa, idMina, idTipoIncidente identificador.Identificador,
	idObra, idReportadoPor *identificador.Identificador,
	fecha, turno, severidad, descripcion, accionInmediata string) (Incidente, error) {
	if !EsValido(turno, TurnosValidos) {
		return Incidente{}, ErrTurnoNoReconocido
	}
	if !EsValido(severidad, SeveridadesValidas) {
		return Incidente{}, ErrSeveridadNoReconocida
	}
	texto := strings.TrimSpace(descripcion)
	if texto == "" {
		return Incidente{}, ErrDescripcionObligatoria
	}
	dia, err := time.Parse("2006-01-02", strings.TrimSpace(fecha))
	if err != nil {
		return Incidente{}, ErrFechaInvalida
	}
	hoy := time.Now()
	if dia.After(time.Date(hoy.Year(), hoy.Month(), hoy.Day(), 23, 59, 59, 0, dia.Location())) {
		return Incidente{}, ErrFechaEnElFuturo
	}
	return Incidente{
		id:              identificador.Nuevo(),
		idEmpresa:       idEmpresa,
		idMina:          idMina,
		idObra:          idObra,
		idTipoIncidente: idTipoIncidente,
		idReportadoPor:  idReportadoPor,
		fecha:           dia,
		turno:           turno,
		severidad:       severidad,
		descripcion:     texto,
		accionInmediata: strings.TrimSpace(accionInmediata),
	}, nil
}

func (incidente Incidente) Identificador() identificador.Identificador { return incidente.id }
func (incidente Incidente) Empresa() identificador.Identificador       { return incidente.idEmpresa }
func (incidente Incidente) Mina() identificador.Identificador          { return incidente.idMina }
func (incidente Incidente) Obra() *identificador.Identificador         { return incidente.idObra }
func (incidente Incidente) TipoDeIncidente() identificador.Identificador {
	return incidente.idTipoIncidente
}
func (incidente Incidente) ReportadoPor() *identificador.Identificador {
	return incidente.idReportadoPor
}
func (incidente Incidente) Fecha() time.Time        { return incidente.fecha }
func (incidente Incidente) Turno() string           { return incidente.turno }
func (incidente Incidente) Severidad() string       { return incidente.severidad }
func (incidente Incidente) Descripcion() string     { return incidente.descripcion }
func (incidente Incidente) AccionInmediata() string { return incidente.accionInmediata }
