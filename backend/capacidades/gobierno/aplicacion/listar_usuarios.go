package aplicacion

import (
	"context"

	"minas/capacidades/gobierno/puertos"
)

type ListarUsuarios struct {
	unidad puertos.UnidadDeTrabajo
	lector puertos.LectorDeGobierno
}

func NuevoListarUsuarios(unidad puertos.UnidadDeTrabajo, lector puertos.LectorDeGobierno) *ListarUsuarios {
	return &ListarUsuarios{unidad: unidad, lector: lector}
}

func (caso *ListarUsuarios) Ejecutar(ctx context.Context, filtro puertos.FiltroDeUsuarios) ([]puertos.ResumenUsuario, string, error) {
	var usuarios []puertos.ResumenUsuario
	var siguiente string
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		listados, cursor, err := caso.lector.ListarUsuarios(ctx, filtro)
		usuarios, siguiente = listados, cursor
		return err
	})
	return usuarios, siguiente, err
}

type DetalleDeUsuario struct {
	unidad puertos.UnidadDeTrabajo
	lector puertos.LectorDeGobierno
}

func NuevoDetalleDeUsuario(unidad puertos.UnidadDeTrabajo, lector puertos.LectorDeGobierno) *DetalleDeUsuario {
	return &DetalleDeUsuario{unidad: unidad, lector: lector}
}

type FichaDeUsuario struct {
	Usuario      puertos.DetalleUsuario
	Asignaciones []puertos.ResumenAsignacion
}

func (caso *DetalleDeUsuario) Ejecutar(ctx context.Context, identificadorUsuario string) (FichaDeUsuario, bool, error) {
	var ficha FichaDeUsuario
	var encontrado bool
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		detalle, existe, err := caso.lector.DetalleDeUsuario(ctx, identificadorUsuario)
		if err != nil || !existe {
			encontrado = existe
			return err
		}
		asignaciones, err := caso.lector.AsignacionesDe(ctx, identificadorUsuario)
		if err != nil {
			return err
		}
		ficha = FichaDeUsuario{Usuario: detalle, Asignaciones: asignaciones}
		encontrado = true
		return nil
	})
	return ficha, encontrado, err
}
