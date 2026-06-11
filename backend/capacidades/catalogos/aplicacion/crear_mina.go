package aplicacion

import (
	"context"

	"minas/capacidades/catalogos/dominio"
	"minas/capacidades/catalogos/puertos"
	"minas/compartido/identificador"
)

type ComandoCrearMina struct {
	IdentificadorEmpresa string
	Nombre               string
	Area                 string
}

type CrearMina struct {
	unidad puertos.UnidadDeTrabajo
	minas  puertos.RepositorioMina
}

func NuevoCrearMina(unidad puertos.UnidadDeTrabajo, minas puertos.RepositorioMina) *CrearMina {
	return &CrearMina{unidad: unidad, minas: minas}
}

func (caso *CrearMina) Ejecutar(ctx context.Context, comando ComandoCrearMina) (string, error) {
	empresa, err := identificador.Desde(comando.IdentificadorEmpresa)
	if err != nil {
		return "", err
	}
	mina, err := dominio.CrearMina(empresa, comando.Nombre, comando.Area)
	if err != nil {
		return "", err
	}
	err = caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		return caso.minas.Guardar(ctx, mina)
	})
	if err != nil {
		return "", err
	}
	return mina.Identificador().Texto(), nil
}
