package infraestructura

import (
	"context"

	"minas/capacidades/produccion/dominio"
	"minas/compartido/identificador"
	"minas/plataforma/persistencia"
)

type tablasDeCarga struct {
	parte   string
	detalle string
	llave   string
}

func tablasDe(modalidad string) (tablasDeCarga, error) {
	switch modalidad {
	case dominio.ModalidadAcarreo:
		return tablasDeCarga{"produccion.parte_acarreo", "produccion.acarreo_viaje", "id_parte_acarreo"}, nil
	case dominio.ModalidadRezagado:
		return tablasDeCarga{"produccion.parte_rezagado", "produccion.rezagado_ciclo", "id_parte_rezagado"}, nil
	default:
		return tablasDeCarga{}, dominio.ErrModalidadNoReconocida
	}
}

type RepositorioDeCargaPostgres struct{}

func NuevoRepositorioDeCarga() RepositorioDeCargaPostgres {
	return RepositorioDeCargaPostgres{}
}

func (RepositorioDeCargaPostgres) Guardar(ctx context.Context, parte dominio.ParteDeCarga) error {
	tablas, err := tablasDe(parte.Modalidad())
	if err != nil {
		return err
	}
	consultas := persistencia.ConsultasDe(ctx)
	_, err = consultas.Exec(ctx,
		`INSERT INTO `+tablas.parte+` (id, id_empresa, id_mina, id_obra, id_equipo, id_operador, id_supervisor,
		                               fecha, turno, horometro_inicial, horometro_final, observaciones)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NULLIF($12, ''))`,
		parte.Identificador().Texto(), parte.Empresa().Texto(), parte.Mina().Texto(), parte.Obra().Texto(),
		parte.Equipo().Texto(), parte.Operador().Texto(), textoOpcional(parte.Supervisor()),
		parte.Fecha(), parte.Turno(), parte.HorometroInicial(), parte.HorometroFinal(), parte.Observaciones())
	if err != nil {
		return err
	}
	for _, viaje := range parte.Viajes() {
		_, err := consultas.Exec(ctx,
			`INSERT INTO `+tablas.detalle+` (id, id_empresa, `+tablas.llave+`, desde, hasta, id_tipo_mineral, toneladas)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			viaje.Identificador().Texto(), parte.Empresa().Texto(), parte.Identificador().Texto(),
			viaje.Desde(), viaje.Hasta(), viaje.TipoDeMineral().Texto(), viaje.Toneladas())
		if err != nil {
			return err
		}
	}
	return nil
}

type RepositorioDeBarrenacionPostgres struct{}

func NuevoRepositorioDeBarrenacion() RepositorioDeBarrenacionPostgres {
	return RepositorioDeBarrenacionPostgres{}
}

func (RepositorioDeBarrenacionPostgres) Guardar(ctx context.Context, parte dominio.ParteDeBarrenacion) error {
	consultas := persistencia.ConsultasDe(ctx)
	horometros := parte.HorometrosDelTurno()
	_, err := consultas.Exec(ctx,
		`INSERT INTO produccion.parte_barrenacion (id, id_empresa, tipo_barrenacion, id_mina, id_obra, id_equipo,
		                                           id_capitan_mina, id_supervisor, id_operador, id_ayudante, id_actividad,
		                                           fecha, turno,
		                                           horometro_diesel_inicial, horometro_diesel_final,
		                                           horometro_electrico_inicial, horometro_electrico_final,
		                                           horometro_percusion_inicial, horometro_percusion_final,
		                                           observaciones)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, NULLIF($20, ''))`,
		parte.Identificador().Texto(), parte.Empresa().Texto(), parte.TipoDeBarrenacion(), parte.Mina().Texto(),
		parte.Obra().Texto(), parte.Equipo().Texto(), parte.CapitanDeMina().Texto(),
		textoOpcional(parte.Supervisor()), parte.Operador().Texto(), textoOpcional(parte.Ayudante()),
		textoOpcional(parte.Actividad()), parte.Fecha(), parte.Turno(),
		horometros.DieselInicial, horometros.DieselFinal,
		horometros.ElectricoInicial, horometros.ElectricoFinal,
		horometros.PercusionInicial, horometros.PercusionFinal,
		parte.Observaciones())
	if err != nil {
		return err
	}
	for _, avance := range parte.Avances() {
		_, err := consultas.Exec(ctx,
			`INSERT INTO produccion.barrenacion_avance (id, id_empresa, id_parte_barrenacion, actividad, lugar,
			                                            no_secciones, longitud, no_barrenos)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			avance.Identificador().Texto(), parte.Empresa().Texto(), parte.Identificador().Texto(),
			avance.Actividad(), avance.Lugar(), avance.NumeroDeSecciones(), avance.Longitud(), avance.NumeroDeBarrenos())
		if err != nil {
			return err
		}
		for _, barreno := range avance.Ejecutados() {
			_, err := consultas.Exec(ctx,
				`INSERT INTO produccion.barrenacion_ejecutado (id_empresa, id_barrenacion_avance, id_tipo_barreno)
				 VALUES ($1, $2, $3)`,
				parte.Empresa().Texto(), avance.Identificador().Texto(), barreno.Texto())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type RepositorioDeDemoraPostgres struct{}

func NuevoRepositorioDeDemora() RepositorioDeDemoraPostgres {
	return RepositorioDeDemoraPostgres{}
}

func (RepositorioDeDemoraPostgres) Guardar(ctx context.Context, demora dominio.Demora) error {
	consultas := persistencia.ConsultasDe(ctx)
	_, err := consultas.Exec(ctx,
		`INSERT INTO produccion.demora_equipo (id, id_empresa, id_mina, id_equipo, id_tipo_demora,
		                                       fecha, turno, minutos, observacion)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULLIF($9, ''))`,
		demora.Identificador().Texto(), demora.Empresa().Texto(), demora.Mina().Texto(), demora.Equipo().Texto(),
		demora.TipoDeDemora().Texto(), demora.Fecha(), demora.Turno(), demora.Minutos(), demora.Observacion())
	return err
}

func textoOpcional(valor *identificador.Identificador) any {
	if valor == nil {
		return nil
	}
	return valor.Texto()
}
