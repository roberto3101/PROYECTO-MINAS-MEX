package infraestructura

import (
	"context"

	"minas/capacidades/catalogos/contrato"
	"minas/capacidades/catalogos/puertos"
	"minas/plataforma/persistencia"
)

type ServicioCatalogosPostgres struct {
	unidad puertos.UnidadDeTrabajo
}

func NuevoServicioCatalogos(unidad puertos.UnidadDeTrabajo) ServicioCatalogosPostgres {
	return ServicioCatalogosPostgres{unidad: unidad}
}

func (servicio ServicioCatalogosPostgres) MinasActivas(ctx context.Context) ([]contrato.MinaPublicada, error) {
	var minas []contrato.MinaPublicada
	err := servicio.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		consultas := persistencia.ConsultasDe(ctx)
		filas, err := consultas.Query(ctx,
			`SELECT id, nombre FROM catalogos.mina
			 WHERE eliminado_en IS NULL AND estado = 'ACTIVA' ORDER BY nombre`)
		if err != nil {
			return err
		}
		defer filas.Close()
		for filas.Next() {
			var mina contrato.MinaPublicada
			if err := filas.Scan(&mina.Identificador, &mina.Nombre); err != nil {
				return err
			}
			minas = append(minas, mina)
		}
		return filas.Err()
	})
	return minas, err
}
