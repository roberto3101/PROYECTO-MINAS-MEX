package aplicacion

import (
	"context"

	"minas/capacidades/catalogos/dominio"
	"minas/capacidades/catalogos/puertos"
	"minas/compartido/identificador"
)

type ComandoCambiarEstadoDeEmpleado struct {
	IdentificadorEmpleado string
	Estado                string
}

type CambiarEstadoDeEmpleado struct {
	unidad    puertos.UnidadDeTrabajo
	empleados puertos.RepositorioEmpleado
}

func NuevoCambiarEstadoDeEmpleado(unidad puertos.UnidadDeTrabajo, empleados puertos.RepositorioEmpleado) *CambiarEstadoDeEmpleado {
	return &CambiarEstadoDeEmpleado{unidad: unidad, empleados: empleados}
}

func (caso *CambiarEstadoDeEmpleado) Ejecutar(ctx context.Context, comando ComandoCambiarEstadoDeEmpleado) error {
	idEmpleado, err := identificador.Desde(comando.IdentificadorEmpleado)
	if err != nil {
		return err
	}
	if !dominio.EsEstadoValido(comando.Estado, dominio.EstadosDeEmpleadoValidos) {
		return dominio.ErrEstadoNoReconocido
	}
	return caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		return caso.empleados.CambiarEstado(ctx, idEmpleado, comando.Estado)
	})
}
