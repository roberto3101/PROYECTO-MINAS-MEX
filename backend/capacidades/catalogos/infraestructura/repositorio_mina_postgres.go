package infraestructura

import (
	"context"

	"minas/capacidades/catalogos/dominio"
	"minas/plataforma/persistencia"
)

type RepositorioMinaPostgres struct{}

func NuevoRepositorioMina() RepositorioMinaPostgres {
	return RepositorioMinaPostgres{}
}

func (RepositorioMinaPostgres) Guardar(ctx context.Context, mina dominio.Mina) error {
	consultas := persistencia.ConsultasDe(ctx)
	_, err := consultas.Exec(ctx,
		`INSERT INTO catalogos.mina (id, id_empresa, nombre, area)
		 VALUES ($1, $2, $3, NULLIF($4, ''))`,
		mina.Identificador().Texto(), mina.Empresa().Texto(), mina.Nombre(), mina.Area())
	return traducirErrorDeEscritura(err)
}
