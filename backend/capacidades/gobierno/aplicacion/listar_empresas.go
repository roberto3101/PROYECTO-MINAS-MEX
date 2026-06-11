package aplicacion

import (
	"context"

	"minas/capacidades/gobierno/puertos"
)

type ListarEmpresas struct {
	unidad puertos.UnidadDeTrabajoDePlataforma
	lector puertos.LectorDePlataforma
}

func NuevoListarEmpresas(unidad puertos.UnidadDeTrabajoDePlataforma, lector puertos.LectorDePlataforma) *ListarEmpresas {
	return &ListarEmpresas{unidad: unidad, lector: lector}
}

func (caso *ListarEmpresas) Ejecutar(ctx context.Context, filtro puertos.FiltroDeEmpresas) ([]puertos.ResumenEmpresaDePlataforma, string, error) {
	var empresas []puertos.ResumenEmpresaDePlataforma
	var siguiente string
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		listadas, cursor, err := caso.lector.ListarEmpresas(ctx, filtro)
		empresas, siguiente = listadas, cursor
		return err
	})
	return empresas, siguiente, err
}

type DetalleDeEmpresa struct {
	unidad puertos.UnidadDeTrabajoDePlataforma
	lector puertos.LectorDePlataforma
}

func NuevoDetalleDeEmpresa(unidad puertos.UnidadDeTrabajoDePlataforma, lector puertos.LectorDePlataforma) *DetalleDeEmpresa {
	return &DetalleDeEmpresa{unidad: unidad, lector: lector}
}

func (caso *DetalleDeEmpresa) Ejecutar(ctx context.Context, identificador string) (puertos.DetalleEmpresaDePlataforma, bool, error) {
	var detalle puertos.DetalleEmpresaDePlataforma
	var encontrada bool
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		resultado, existe, err := caso.lector.DetalleDeEmpresa(ctx, identificador)
		detalle, encontrada = resultado, existe
		return err
	})
	return detalle, encontrada, err
}
