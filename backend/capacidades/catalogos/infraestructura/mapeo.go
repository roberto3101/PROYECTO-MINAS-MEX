package infraestructura

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	"minas/capacidades/catalogos/dominio"
	"minas/compartido/identificador"
)

const codigoPostgresLlaveDuplicada = "23505"

func traducirErrorDeEscritura(err error) error {
	var errorPostgres *pgconn.PgError
	if errors.As(err, &errorPostgres) && errorPostgres.Code == codigoPostgresLlaveDuplicada {
		return dominio.ErrYaExiste
	}
	return err
}

func textoOpcional(valor *identificador.Identificador) *string {
	if valor == nil {
		return nil
	}
	texto := valor.Texto()
	return &texto
}
