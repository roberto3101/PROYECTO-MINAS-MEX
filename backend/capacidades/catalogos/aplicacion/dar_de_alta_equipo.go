package aplicacion

import (
	"context"

	"minas/capacidades/catalogos/dominio"
	"minas/capacidades/catalogos/puertos"
	"minas/compartido/identificador"
)

type ComandoDarDeAltaEquipo struct {
	IdentificadorEmpresa       string
	IdentificadorMina          string
	IdentificadorTipoEquipo    string
	IdentificadorModuloTrabajo string
	Codigo                     string
	Fabricante                 string
}

type DarDeAltaEquipo struct {
	unidad  puertos.UnidadDeTrabajo
	equipos puertos.RepositorioEquipo
}

func NuevoDarDeAltaEquipo(unidad puertos.UnidadDeTrabajo, equipos puertos.RepositorioEquipo) *DarDeAltaEquipo {
	return &DarDeAltaEquipo{unidad: unidad, equipos: equipos}
}

func (caso *DarDeAltaEquipo) Ejecutar(ctx context.Context, comando ComandoDarDeAltaEquipo) (string, error) {
	empresa, err := identificador.Desde(comando.IdentificadorEmpresa)
	if err != nil {
		return "", err
	}
	mina, err := identificador.Desde(comando.IdentificadorMina)
	if err != nil {
		return "", err
	}
	tipoEquipo, err := identificador.Desde(comando.IdentificadorTipoEquipo)
	if err != nil {
		return "", err
	}
	moduloTrabajo, err := identificador.Desde(comando.IdentificadorModuloTrabajo)
	if err != nil {
		return "", err
	}
	equipo, err := dominio.DarDeAltaEquipo(empresa, mina, tipoEquipo, moduloTrabajo, comando.Codigo, comando.Fabricante)
	if err != nil {
		return "", err
	}
	err = caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		return caso.equipos.Guardar(ctx, equipo)
	})
	if err != nil {
		return "", err
	}
	return equipo.Identificador().Texto(), nil
}
