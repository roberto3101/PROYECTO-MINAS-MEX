package aplicacion

import (
	"context"

	"minas/capacidades/catalogos/puertos"
)

type ListarTiposDeObra struct {
	unidad puertos.UnidadDeTrabajo
	lector puertos.LectorDeCatalogos
}

func NuevoListarTiposDeObra(unidad puertos.UnidadDeTrabajo, lector puertos.LectorDeCatalogos) *ListarTiposDeObra {
	return &ListarTiposDeObra{unidad: unidad, lector: lector}
}

func (caso *ListarTiposDeObra) Ejecutar(ctx context.Context) ([]puertos.OpcionDeCatalogo, error) {
	var opciones []puertos.OpcionDeCatalogo
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		listadas, err := caso.lector.ListarTiposDeObra(ctx)
		opciones = listadas
		return err
	})
	return opciones, err
}

type ListarTiposDeMineral struct {
	unidad puertos.UnidadDeTrabajo
	lector puertos.LectorDeCatalogos
}

func NuevoListarTiposDeMineral(unidad puertos.UnidadDeTrabajo, lector puertos.LectorDeCatalogos) *ListarTiposDeMineral {
	return &ListarTiposDeMineral{unidad: unidad, lector: lector}
}

func (caso *ListarTiposDeMineral) Ejecutar(ctx context.Context) ([]puertos.OpcionDeCatalogo, error) {
	var opciones []puertos.OpcionDeCatalogo
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		listadas, err := caso.lector.ListarTiposDeMineral(ctx)
		opciones = listadas
		return err
	})
	return opciones, err
}

type ListarTiposDeBarreno struct {
	unidad puertos.UnidadDeTrabajo
	lector puertos.LectorDeCatalogos
}

func NuevoListarTiposDeBarreno(unidad puertos.UnidadDeTrabajo, lector puertos.LectorDeCatalogos) *ListarTiposDeBarreno {
	return &ListarTiposDeBarreno{unidad: unidad, lector: lector}
}

func (caso *ListarTiposDeBarreno) Ejecutar(ctx context.Context) ([]puertos.OpcionDeCatalogo, error) {
	var opciones []puertos.OpcionDeCatalogo
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		listadas, err := caso.lector.ListarTiposDeBarreno(ctx)
		opciones = listadas
		return err
	})
	return opciones, err
}

type ListarTiposDeDemora struct {
	unidad puertos.UnidadDeTrabajo
	lector puertos.LectorDeCatalogos
}

func NuevoListarTiposDeDemora(unidad puertos.UnidadDeTrabajo, lector puertos.LectorDeCatalogos) *ListarTiposDeDemora {
	return &ListarTiposDeDemora{unidad: unidad, lector: lector}
}

func (caso *ListarTiposDeDemora) Ejecutar(ctx context.Context) ([]puertos.OpcionDeCatalogo, error) {
	var opciones []puertos.OpcionDeCatalogo
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		listadas, err := caso.lector.ListarTiposDeDemora(ctx)
		opciones = listadas
		return err
	})
	return opciones, err
}
