package aplicacion

import (
	"context"

	contratoCatalogos "minas/capacidades/catalogos/contrato"
	"minas/capacidades/gobierno/puertos"
	"minas/compartido/identificador"
	"minas/plataforma/contexto"
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
	unidad    puertos.UnidadDeTrabajoDePlataforma
	lector    puertos.LectorDePlataforma
	catalogos contratoCatalogos.Catalogos
}

func NuevoDetalleDeEmpresa(unidad puertos.UnidadDeTrabajoDePlataforma, lector puertos.LectorDePlataforma, catalogos contratoCatalogos.Catalogos) *DetalleDeEmpresa {
	return &DetalleDeEmpresa{unidad: unidad, lector: lector, catalogos: catalogos}
}

func (caso *DetalleDeEmpresa) Ejecutar(ctx context.Context, identificadorEmpresa string) (puertos.DetalleEmpresaDePlataforma, bool, error) {
	var detalle puertos.DetalleEmpresaDePlataforma
	var encontrada bool
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		resultado, existe, err := caso.lector.DetalleDeEmpresa(ctx, identificadorEmpresa)
		detalle, encontrada = resultado, existe
		return err
	})
	if err != nil || !encontrada {
		return detalle, encontrada, err
	}
	empresa, err := identificador.Desde(detalle.Identificador)
	if err != nil {
		return detalle, true, err
	}
	total, err := caso.catalogos.TotalDeMinas(contexto.ConEmpresaImpersonada(ctx, empresa))
	if err != nil {
		return detalle, true, err
	}
	detalle.TotalMinas = total
	return detalle, true, nil
}
