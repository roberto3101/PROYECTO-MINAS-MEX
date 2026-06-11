package persistencia

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UnidadDePlataformaPostgres struct {
	pool *pgxpool.Pool
}

func NuevaUnidadDePlataforma(pool *pgxpool.Pool) *UnidadDePlataformaPostgres {
	return &UnidadDePlataformaPostgres{pool: pool}
}

func (unidad *UnidadDePlataformaPostgres) EnTransaccion(ctx context.Context, operacion func(ctx context.Context) error) error {
	transaccion, err := unidad.pool.Begin(ctx)
	if err != nil {
		return err
	}
	if _, err := transaccion.Exec(ctx, "SET LOCAL ROLE plataforma"); err != nil {
		_ = transaccion.Rollback(ctx)
		return err
	}
	if err := operacion(conConsultas(ctx, transaccion)); err != nil {
		_ = transaccion.Rollback(ctx)
		return err
	}
	return transaccion.Commit(ctx)
}
