package aplicacion

import (
	"context"

	"minas/capacidades/catalogos/dominio"
	"minas/capacidades/catalogos/puertos"
	"minas/compartido/identificador"
)

type ComandoContratarEmpleado struct {
	IdentificadorEmpresa string
	IdentificadorMina    string
	NumeroNomina         string
	NombreCompleto       string
}

type ContratarEmpleado struct {
	unidad    puertos.UnidadDeTrabajo
	empleados puertos.RepositorioEmpleado
}

func NuevoContratarEmpleado(unidad puertos.UnidadDeTrabajo, empleados puertos.RepositorioEmpleado) *ContratarEmpleado {
	return &ContratarEmpleado{unidad: unidad, empleados: empleados}
}

func (caso *ContratarEmpleado) Ejecutar(ctx context.Context, comando ComandoContratarEmpleado) (string, error) {
	empresa, err := identificador.Desde(comando.IdentificadorEmpresa)
	if err != nil {
		return "", err
	}
	mina, err := identificador.Desde(comando.IdentificadorMina)
	if err != nil {
		return "", err
	}
	empleado, err := dominio.ContratarEmpleado(empresa, mina, comando.NumeroNomina, comando.NombreCompleto)
	if err != nil {
		return "", err
	}
	err = caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		return caso.empleados.Guardar(ctx, empleado)
	})
	if err != nil {
		return "", err
	}
	return empleado.Identificador().Texto(), nil
}
