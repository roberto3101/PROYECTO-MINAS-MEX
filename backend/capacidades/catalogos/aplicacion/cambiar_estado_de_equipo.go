package aplicacion

import (
	"context"

	"minas/capacidades/catalogos/dominio"
	"minas/capacidades/catalogos/puertos"
	"minas/compartido/identificador"
)

type ComandoCambiarEstadoDeEquipo struct {
	IdentificadorEquipo string
	Estado              string
}

type CambiarEstadoDeEquipo struct {
	unidad  puertos.UnidadDeTrabajo
	equipos puertos.RepositorioEquipo
}

func NuevoCambiarEstadoDeEquipo(unidad puertos.UnidadDeTrabajo, equipos puertos.RepositorioEquipo) *CambiarEstadoDeEquipo {
	return &CambiarEstadoDeEquipo{unidad: unidad, equipos: equipos}
}

func (caso *CambiarEstadoDeEquipo) Ejecutar(ctx context.Context, comando ComandoCambiarEstadoDeEquipo) error {
	idEquipo, err := identificador.Desde(comando.IdentificadorEquipo)
	if err != nil {
		return err
	}
	if !dominio.EsEstadoValido(comando.Estado, dominio.EstadosDeEquipoValidos) {
		return dominio.ErrEstadoNoReconocido
	}
	return caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		return caso.equipos.CambiarEstado(ctx, idEquipo, comando.Estado)
	})
}
