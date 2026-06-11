package infraestructura

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"minas/capacidades/gobierno/dominio"
	"minas/compartido/identificador"
	"minas/plataforma/persistencia"
)

type RepositorioPermisoPostgres struct{}

func NuevoRepositorioPermiso() RepositorioPermisoPostgres {
	return RepositorioPermisoPostgres{}
}

func (RepositorioPermisoPostgres) BuscarPorCodigo(ctx context.Context, codigo string) (dominio.Permiso, bool, error) {
	consultas := persistencia.ConsultasDe(ctx)
	var id, codigoPermiso, descripcion, modulo string
	fila := consultas.QueryRow(ctx,
		"SELECT id, codigo, descripcion, modulo FROM gobierno.permiso WHERE codigo = $1 AND eliminado_en IS NULL",
		codigo)
	if err := fila.Scan(&id, &codigoPermiso, &descripcion, &modulo); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dominio.Permiso{}, false, nil
		}
		return dominio.Permiso{}, false, err
	}
	convertido, err := identificador.Desde(id)
	if err != nil {
		return dominio.Permiso{}, false, err
	}
	return dominio.ReconstruirPermiso(convertido, codigoPermiso, descripcion, modulo), true, nil
}

func (RepositorioPermisoPostgres) Listar(ctx context.Context) ([]dominio.Permiso, error) {
	consultas := persistencia.ConsultasDe(ctx)
	filas, err := consultas.Query(ctx,
		"SELECT id, codigo, descripcion, modulo FROM gobierno.permiso WHERE eliminado_en IS NULL ORDER BY modulo, codigo")
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	var catalogo []dominio.Permiso
	for filas.Next() {
		var id, codigo, descripcion, modulo string
		if err := filas.Scan(&id, &codigo, &descripcion, &modulo); err != nil {
			return nil, err
		}
		convertido, err := identificador.Desde(id)
		if err != nil {
			return nil, err
		}
		catalogo = append(catalogo, dominio.ReconstruirPermiso(convertido, codigo, descripcion, modulo))
	}
	return catalogo, filas.Err()
}
