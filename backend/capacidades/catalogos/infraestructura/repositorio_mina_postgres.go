package infraestructura

import (
	"context"

	"minas/capacidades/catalogos/dominio"
	"minas/compartido/identificador"
	"minas/plataforma/persistencia"
)

type RepositorioMinaPostgres struct{}

func NuevoRepositorioMina() RepositorioMinaPostgres {
	return RepositorioMinaPostgres{}
}

func (RepositorioMinaPostgres) Guardar(ctx context.Context, mina dominio.Mina) error {
	consultas := persistencia.ConsultasDe(ctx)
	_, err := consultas.Exec(ctx,
		`INSERT INTO catalogos.mina (id, id_empresa, nombre, area, niveles, densidad_promedio,
		                             seccion_obra_largo, seccion_obra_ancho, seccion_veta_largo, seccion_veta_ancho)
		 VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, $7, $8, $9, $10)`,
		mina.Identificador().Texto(), mina.Empresa().Texto(), mina.Nombre(), mina.Area(), mina.Niveles(),
		mina.DensidadPromedio(), mina.SeccionObraLargo(), mina.SeccionObraAncho(),
		mina.SeccionVetaLargo(), mina.SeccionVetaAncho())
	return traducirErrorDeEscritura(err)
}

func (RepositorioMinaPostgres) CambiarEstado(ctx context.Context, id identificador.Identificador, estado string) error {
	consultas := persistencia.ConsultasDe(ctx)
	_, err := consultas.Exec(ctx,
		`UPDATE catalogos.mina SET estado = $2, actualizado_en = now() WHERE id = $1`,
		id.Texto(), estado)
	return err
}
