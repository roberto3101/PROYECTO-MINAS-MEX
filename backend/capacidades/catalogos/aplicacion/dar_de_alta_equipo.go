package aplicacion

import (
	"context"
	"time"

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
	Descripcion                string
	ModeloPerforadora          string
	CapacidadLongitud          *float64
	Fabricante                 string
	Modelo                     string
	NumeroSerie                string
	AnioFabricacion            *int
	FechaIngresoMina           string
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
	fechaIngreso, err := fechaOpcional(comando.FechaIngresoMina)
	if err != nil {
		return "", err
	}
	equipo, err := dominio.DarDeAltaEquipo(empresa, mina, tipoEquipo, moduloTrabajo,
		comando.Codigo, comando.Descripcion, comando.ModeloPerforadora, comando.Fabricante,
		comando.Modelo, comando.NumeroSerie, comando.CapacidadLongitud, comando.AnioFabricacion, fechaIngreso)
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

func fechaOpcional(texto string) (*time.Time, error) {
	if texto == "" {
		return nil, nil
	}
	fecha, err := time.Parse("2006-01-02", texto)
	if err != nil {
		return nil, dominio.ErrValorFueraDeRango
	}
	return &fecha, nil
}
