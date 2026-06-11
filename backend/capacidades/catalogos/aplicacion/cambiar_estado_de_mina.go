package aplicacion

import (
	"context"

	"minas/capacidades/catalogos/dominio"
	"minas/capacidades/catalogos/puertos"
	"minas/compartido/identificador"
)

type ComandoCambiarEstadoDeMina struct {
	IdentificadorMina string
	Estado            string
}

type CambiarEstadoDeMina struct {
	unidad puertos.UnidadDeTrabajo
	minas  puertos.RepositorioMina
}

func NuevoCambiarEstadoDeMina(unidad puertos.UnidadDeTrabajo, minas puertos.RepositorioMina) *CambiarEstadoDeMina {
	return &CambiarEstadoDeMina{unidad: unidad, minas: minas}
}

func (caso *CambiarEstadoDeMina) Ejecutar(ctx context.Context, comando ComandoCambiarEstadoDeMina) error {
	idMina, err := identificador.Desde(comando.IdentificadorMina)
	if err != nil {
		return err
	}
	if !dominio.EsEstadoValido(comando.Estado, dominio.EstadosDeMinaValidos) {
		return dominio.ErrEstadoNoReconocido
	}
	return caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		return caso.minas.CambiarEstado(ctx, idMina, comando.Estado)
	})
}
