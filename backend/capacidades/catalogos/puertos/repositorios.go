package puertos

import (
	"context"

	"minas/capacidades/catalogos/dominio"
	"minas/compartido/identificador"
)

type RepositorioMina interface {
	Guardar(ctx context.Context, mina dominio.Mina) error
	CambiarEstado(ctx context.Context, id identificador.Identificador, estado string) error
}

type RepositorioEmpleado interface {
	Guardar(ctx context.Context, empleado dominio.Empleado) error
	CambiarEstado(ctx context.Context, id identificador.Identificador, estado string) error
}

type RepositorioEquipo interface {
	Guardar(ctx context.Context, equipo dominio.Equipo) error
	CambiarEstado(ctx context.Context, id identificador.Identificador, estado string) error
}

type UnidadDeTrabajo interface {
	EnTransaccion(ctx context.Context, operacion func(ctx context.Context) error) error
}
