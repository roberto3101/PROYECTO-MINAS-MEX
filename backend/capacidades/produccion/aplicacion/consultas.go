package aplicacion

import (
	"context"

	"minas/capacidades/produccion/puertos"
)

type ConsultasDeProduccion struct {
	unidad puertos.UnidadDeTrabajo
	lector puertos.LectorDeProduccion
}

func NuevasConsultasDeProduccion(unidad puertos.UnidadDeTrabajo, lector puertos.LectorDeProduccion) *ConsultasDeProduccion {
	return &ConsultasDeProduccion{unidad: unidad, lector: lector}
}

func (caso *ConsultasDeProduccion) ListarPartesDeCarga(ctx context.Context, modalidad string, filtro puertos.FiltroDeProduccion) ([]puertos.ResumenDeParte, string, error) {
	var partes []puertos.ResumenDeParte
	var siguiente string
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		listados, cursor, err := caso.lector.ListarPartesDeCarga(ctx, modalidad, filtro)
		partes, siguiente = listados, cursor
		return err
	})
	return partes, siguiente, err
}

func (caso *ConsultasDeProduccion) DetalleDeParteDeCarga(ctx context.Context, modalidad, identificadorParte string) (puertos.DetalleDeParteDeCarga, bool, error) {
	var detalle puertos.DetalleDeParteDeCarga
	var encontrado bool
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		resultado, existe, err := caso.lector.DetalleDeParteDeCarga(ctx, modalidad, identificadorParte)
		detalle, encontrado = resultado, existe
		return err
	})
	return detalle, encontrado, err
}

func (caso *ConsultasDeProduccion) ListarPartesDeBarrenacion(ctx context.Context, filtro puertos.FiltroDeProduccion) ([]puertos.ResumenDeParte, string, error) {
	var partes []puertos.ResumenDeParte
	var siguiente string
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		listados, cursor, err := caso.lector.ListarPartesDeBarrenacion(ctx, filtro)
		partes, siguiente = listados, cursor
		return err
	})
	return partes, siguiente, err
}

func (caso *ConsultasDeProduccion) DetalleDeParteDeBarrenacion(ctx context.Context, identificadorParte string) (puertos.DetalleDeParteDeBarrenacion, bool, error) {
	var detalle puertos.DetalleDeParteDeBarrenacion
	var encontrado bool
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		resultado, existe, err := caso.lector.DetalleDeParteDeBarrenacion(ctx, identificadorParte)
		detalle, encontrado = resultado, existe
		return err
	})
	return detalle, encontrado, err
}

func (caso *ConsultasDeProduccion) ListarDemoras(ctx context.Context, filtro puertos.FiltroDeProduccion) ([]puertos.ResumenDeDemora, string, error) {
	var demoras []puertos.ResumenDeDemora
	var siguiente string
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		listadas, cursor, err := caso.lector.ListarDemoras(ctx, filtro)
		demoras, siguiente = listadas, cursor
		return err
	})
	return demoras, siguiente, err
}

func (caso *ConsultasDeProduccion) Indicadores(ctx context.Context, filtro puertos.FiltroDeProduccion) (puertos.IndicadoresDeProduccion, error) {
	var indicadores puertos.IndicadoresDeProduccion
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		resultado, err := caso.lector.Indicadores(ctx, filtro)
		indicadores = resultado
		return err
	})
	return indicadores, err
}
