package persistencia

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NuevoPool(ctx context.Context, cadenaDeConexion string) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, cadenaDeConexion)
}
