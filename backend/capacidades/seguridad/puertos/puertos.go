package puertos

import (
	"context"

	"minas/capacidades/seguridad/dominio"
)

type UnidadDeTrabajo interface {
	EnTransaccion(ctx context.Context, operacion func(ctx context.Context) error) error
}

type RepositorioDeIncidentes interface {
	Guardar(ctx context.Context, incidente dominio.Incidente) error
}

type FiltroDeIncidentes struct {
	Mina              string
	Severidad         string
	Desde             string
	Hasta             string
	Cursor            string
	Limite            int
	RestringirPorMina bool
	MinasPermitidas   []string
}

type ResumenDeIncidente struct {
	Identificador   string
	Fecha           string
	Turno           string
	Mina            string
	Obra            string
	TipoDeIncidente string
	RequierePar     bool
	Severidad       string
	Descripcion     string
	AccionInmediata string
	ReportadoPor    string
}

type OpcionDeSeguridad struct {
	Identificador string
	Codigo        string
	Descripcion   string
	RequiereParo  bool
}

type ConteoPorSeveridad struct {
	Severidad string
	Total     int
}

type LectorDeSeguridad interface {
	ListarIncidentes(ctx context.Context, filtro FiltroDeIncidentes) ([]ResumenDeIncidente, string, error)
	ListarTiposDeIncidente(ctx context.Context) ([]OpcionDeSeguridad, error)
	ConteoPorSeveridad(ctx context.Context, filtro FiltroDeIncidentes) ([]ConteoPorSeveridad, error)
}
