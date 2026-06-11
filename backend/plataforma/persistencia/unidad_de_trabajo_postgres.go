package persistencia

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	"minas/plataforma/contexto"
)

var ErrSinTenant = errors.New("la operacion requiere un tenant en el contexto")

type UnidadDeTrabajoPostgres struct {
	pool *pgxpool.Pool
}

func NuevaUnidadDeTrabajo(pool *pgxpool.Pool) *UnidadDeTrabajoPostgres {
	return &UnidadDeTrabajoPostgres{pool: pool}
}

func (unidad *UnidadDeTrabajoPostgres) EnTransaccion(ctx context.Context, operacion func(ctx context.Context) error) error {
	tenant, presente := contexto.TenantDe(ctx)
	if !presente {
		return ErrSinTenant
	}
	transaccion, err := unidad.pool.Begin(ctx)
	if err != nil {
		return err
	}
	if _, err := transaccion.Exec(ctx, "SELECT set_config('app.empresa_actual', $1, true)", tenant.Empresa.Texto()); err != nil {
		_ = transaccion.Rollback(ctx)
		return err
	}
	if _, err := transaccion.Exec(ctx, "SET LOCAL ROLE "+nombreDeRolSeguro(tenant.Rol)); err != nil {
		_ = transaccion.Rollback(ctx)
		return err
	}
	if err := operacion(conConsultas(ctx, transaccion)); err != nil {
		_ = transaccion.Rollback(ctx)
		return err
	}
	return transaccion.Commit(ctx)
}

func nombreDeRolSeguro(rol contexto.RolDeBaseDeDatos) string {
	if rol == contexto.RolPlataforma {
		return string(contexto.RolPlataforma)
	}
	return string(contexto.RolAplicacion)
}
