package puertos

import (
	"context"

	"minas/capacidades/catalogos/dominio"
)

type RepositorioMina interface {
	Guardar(ctx context.Context, mina dominio.Mina) error
}

type RepositorioEmpleado interface {
	Guardar(ctx context.Context, empleado dominio.Empleado) error
}

type RepositorioEquipo interface {
	Guardar(ctx context.Context, equipo dominio.Equipo) error
}

type UnidadDeTrabajo interface {
	EnTransaccion(ctx context.Context, operacion func(ctx context.Context) error) error
}
