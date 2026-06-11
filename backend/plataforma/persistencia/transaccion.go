package persistencia

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Consultas interface {
	Exec(ctx context.Context, instruccion string, argumentos ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, instruccion string, argumentos ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, instruccion string, argumentos ...any) pgx.Row
}

type clave int

const claveConsultas clave = iota

func conConsultas(ctx context.Context, consultas Consultas) context.Context {
	return context.WithValue(ctx, claveConsultas, consultas)
}

func ConsultasDe(ctx context.Context) Consultas {
	return ctx.Value(claveConsultas).(Consultas)
}
