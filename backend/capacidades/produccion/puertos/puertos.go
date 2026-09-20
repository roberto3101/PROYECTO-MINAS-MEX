package puertos

import (
	"context"

	"minas/capacidades/produccion/dominio"
)

type UnidadDeTrabajo interface {
	EnTransaccion(ctx context.Context, operacion func(ctx context.Context) error) error
}

type RepositorioDeCarga interface {
	Guardar(ctx context.Context, parte dominio.ParteDeCarga) error
}

type RepositorioDeBarrenacion interface {
	Guardar(ctx context.Context, parte dominio.ParteDeBarrenacion) error
}

type RepositorioDeDemora interface {
	Guardar(ctx context.Context, demora dominio.Demora) error
}

type FiltroDeProduccion struct {
	Mina              string
	Obra              string
	Equipo            string
	Turno             string
	Desde             string
	Hasta             string
	Cursor            string
	Limite            int
	RestringirPorMina bool
	MinasPermitidas   []string
}

type ResumenDeParte struct {
	Identificador string
	Fecha         string
	Turno         string
	Mina          string
	Obra          string
	Equipo        string
	Operador      string
	Horas         string
	Toneladas     string
	Detalles      int
}

type LineaDeCarga struct {
	Identificador string
	Desde         string
	Hasta         string
	TipoDeMineral string
	Toneladas     string
}

type DetalleDeParteDeCarga struct {
	ResumenDeParte
	Observaciones    string
	HorometroInicial string
	HorometroFinal   string
	Viajes           []LineaDeCarga
}

type LineaDeAvance struct {
	Identificador     string
	Actividad         string
	Lugar             string
	NumeroDeSecciones string
	Longitud          string
	NumeroDeBarrenos  int
	Ejecutados        []string
}

type DetalleDeParteDeBarrenacion struct {
	ResumenDeParte
	TipoDeBarrenacion string
	CapitanDeMina     string
	Observaciones     string
	Avances           []LineaDeAvance
}

type ResumenDeDemora struct {
	Identificador string
	Fecha         string
	Turno         string
	Mina          string
	Equipo        string
	TipoDeDemora  string
	Minutos       int
	Observacion   string
}

type IndicadoresDeProduccion struct {
	PartesDeAcarreo     int
	PartesDeRezagado    int
	PartesDeBarrenacion int
	ToneladasAcarreadas string
	ToneladasRezagadas  string
	MetrosBarrenados    string
	MinutosDeDemora     int
}

type LectorDeProduccion interface {
	ListarPartesDeCarga(ctx context.Context, modalidad string, filtro FiltroDeProduccion) ([]ResumenDeParte, string, error)
	DetalleDeParteDeCarga(ctx context.Context, modalidad, identificadorParte string) (DetalleDeParteDeCarga, bool, error)
	ListarPartesDeBarrenacion(ctx context.Context, filtro FiltroDeProduccion) ([]ResumenDeParte, string, error)
	DetalleDeParteDeBarrenacion(ctx context.Context, identificadorParte string) (DetalleDeParteDeBarrenacion, bool, error)
	ListarDemoras(ctx context.Context, filtro FiltroDeProduccion) ([]ResumenDeDemora, string, error)
	Indicadores(ctx context.Context, filtro FiltroDeProduccion) (IndicadoresDeProduccion, error)
}
