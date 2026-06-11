package puertos

import "context"

type UnidadDeTrabajo interface {
	EnTransaccion(ctx context.Context, operacion func(ctx context.Context) error) error
}
