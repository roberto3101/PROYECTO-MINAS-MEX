package infraestructura

import (
	"context"

	"minas/capacidades/produccion/puertos"
	"minas/compartido/paginacion"
	"minas/plataforma/persistencia"
)

type LectorDeProduccionPostgres struct{}

func NuevoLectorDeProduccion() LectorDeProduccionPostgres {
	return LectorDeProduccionPostgres{}
}

func (LectorDeProduccionPostgres) ListarPartesDeCarga(ctx context.Context, modalidad string, filtro puertos.FiltroDeProduccion) ([]puertos.ResumenDeParte, string, error) {
	tablas, err := tablasDe(modalidad)
	if err != nil {
		return nil, "", err
	}
	consultas := persistencia.ConsultasDe(ctx)
	orden, identificadorCursor, err := paginacion.DecodificarCursor(filtro.Cursor)
	if err != nil {
		return nil, "", err
	}
	filas, err := consultas.Query(ctx,
		`SELECT p.id, to_char(p.fecha, 'YYYY-MM-DD'), p.turno, m.nombre, o.codigo,
		        COALESCE(eq.codigo, ''), op.nombre_completo,
		        to_char(p.horometro_final - p.horometro_inicial, 'FM999999990.00'),
		        COALESCE((SELECT to_char(sum(d.toneladas), 'FM999999990.000') FROM `+tablas.detalle+` d
		                  WHERE d.`+tablas.llave+` = p.id), '0.000'),
		        (SELECT count(*) FROM `+tablas.detalle+` d WHERE d.`+tablas.llave+` = p.id)
		 FROM `+tablas.parte+` p
		 JOIN catalogos.mina m ON m.id = p.id_mina
		 JOIN catalogos.obra o ON o.id = p.id_obra
		 JOIN catalogos.equipo eq ON eq.id = p.id_equipo
		 JOIN catalogos.empleado op ON op.id = p.id_operador
		 WHERE p.eliminado_en IS NULL
		   AND ($1 = '' OR p.id_mina::text = $1)
		   AND ($2 = '' OR p.id_obra::text = $2)
		   AND ($3 = '' OR p.id_equipo::text = $3)
		   AND ($4 = '' OR p.turno = $4)
		   AND ($5 = '' OR p.fecha >= $5::date)
		   AND ($6 = '' OR p.fecha <= $6::date)
		   AND (NOT $9 OR p.id_mina = ANY($10::uuid[]))
		   AND ($7 = '' OR (to_char(p.fecha, 'YYYY-MM-DD'), p.id::text) > ($7, $8))
		 ORDER BY p.fecha, p.id LIMIT $11`,
		filtro.Mina, filtro.Obra, filtro.Equipo, filtro.Turno, filtro.Desde, filtro.Hasta,
		orden, identificadorCursor, filtro.RestringirPorMina, filtro.MinasPermitidas, filtro.Limite+1)
	if err != nil {
		return nil, "", err
	}
	defer filas.Close()
	partes, err := leerResumenes(filas)
	if err != nil {
		return nil, "", err
	}
	siguiente := ""
	if len(partes) > filtro.Limite {
		partes = partes[:filtro.Limite]
		ultimo := partes[len(partes)-1]
		siguiente = paginacion.CodificarCursor(ultimo.Fecha, ultimo.Identificador)
	}
	return partes, siguiente, nil
}

func (LectorDeProduccionPostgres) DetalleDeParteDeCarga(ctx context.Context, modalidad, identificadorParte string) (puertos.DetalleDeParteDeCarga, bool, error) {
	tablas, err := tablasDe(modalidad)
	if err != nil {
		return puertos.DetalleDeParteDeCarga{}, false, err
	}
	consultas := persistencia.ConsultasDe(ctx)
	var detalle puertos.DetalleDeParteDeCarga
	fila := consultas.QueryRow(ctx,
		`SELECT p.id, to_char(p.fecha, 'YYYY-MM-DD'), p.turno, m.nombre, o.codigo,
		        COALESCE(eq.codigo, ''), op.nombre_completo,
		        to_char(p.horometro_final - p.horometro_inicial, 'FM999999990.00'),
		        COALESCE((SELECT to_char(sum(d.toneladas), 'FM999999990.000') FROM `+tablas.detalle+` d
		                  WHERE d.`+tablas.llave+` = p.id), '0.000'),
		        (SELECT count(*) FROM `+tablas.detalle+` d WHERE d.`+tablas.llave+` = p.id),
		        COALESCE(p.observaciones, ''),
		        to_char(p.horometro_inicial, 'FM999999990.00'), to_char(p.horometro_final, 'FM999999990.00')
		 FROM `+tablas.parte+` p
		 JOIN catalogos.mina m ON m.id = p.id_mina
		 JOIN catalogos.obra o ON o.id = p.id_obra
		 JOIN catalogos.equipo eq ON eq.id = p.id_equipo
		 JOIN catalogos.empleado op ON op.id = p.id_operador
		 WHERE p.id = $1 AND p.eliminado_en IS NULL`, identificadorParte)
	if err := fila.Scan(&detalle.Identificador, &detalle.Fecha, &detalle.Turno, &detalle.Mina, &detalle.Obra,
		&detalle.Equipo, &detalle.Operador, &detalle.Horas, &detalle.Toneladas, &detalle.Detalles,
		&detalle.Observaciones, &detalle.HorometroInicial, &detalle.HorometroFinal); err != nil {
		if err.Error() == "no rows in result set" {
			return puertos.DetalleDeParteDeCarga{}, false, nil
		}
		return puertos.DetalleDeParteDeCarga{}, false, err
	}
	filas, err := consultas.Query(ctx,
		`SELECT d.id, d.desde, d.hasta, tm.descripcion, COALESCE(to_char(d.toneladas, 'FM999999990.000'), '')
		 FROM `+tablas.detalle+` d
		 JOIN catalogos.tipo_mineral tm ON tm.id = d.id_tipo_mineral
		 WHERE d.`+tablas.llave+` = $1 ORDER BY d.creado_en`, identificadorParte)
	if err != nil {
		return puertos.DetalleDeParteDeCarga{}, false, err
	}
	defer filas.Close()
	for filas.Next() {
		var linea puertos.LineaDeCarga
		if err := filas.Scan(&linea.Identificador, &linea.Desde, &linea.Hasta, &linea.TipoDeMineral, &linea.Toneladas); err != nil {
			return puertos.DetalleDeParteDeCarga{}, false, err
		}
		detalle.Viajes = append(detalle.Viajes, linea)
	}
	return detalle, true, filas.Err()
}

func (LectorDeProduccionPostgres) ListarPartesDeBarrenacion(ctx context.Context, filtro puertos.FiltroDeProduccion) ([]puertos.ResumenDeParte, string, error) {
	consultas := persistencia.ConsultasDe(ctx)
	orden, identificadorCursor, err := paginacion.DecodificarCursor(filtro.Cursor)
	if err != nil {
		return nil, "", err
	}
	filas, err := consultas.Query(ctx,
		`SELECT p.id, to_char(p.fecha, 'YYYY-MM-DD'), p.turno, m.nombre, o.codigo,
		        COALESCE(eq.codigo, ''), op.nombre_completo,
		        COALESCE(to_char(p.horometro_percusion_final - p.horometro_percusion_inicial, 'FM999999990.00'), '0.00'),
		        COALESCE((SELECT to_char(sum(a.longitud), 'FM999999990.00') FROM produccion.barrenacion_avance a
		                  WHERE a.id_parte_barrenacion = p.id), '0.00'),
		        (SELECT count(*) FROM produccion.barrenacion_avance a WHERE a.id_parte_barrenacion = p.id)
		 FROM produccion.parte_barrenacion p
		 JOIN catalogos.mina m ON m.id = p.id_mina
		 JOIN catalogos.obra o ON o.id = p.id_obra
		 JOIN catalogos.equipo eq ON eq.id = p.id_equipo
		 JOIN catalogos.empleado op ON op.id = p.id_operador
		 WHERE p.eliminado_en IS NULL
		   AND ($1 = '' OR p.id_mina::text = $1)
		   AND ($2 = '' OR p.id_obra::text = $2)
		   AND ($3 = '' OR p.id_equipo::text = $3)
		   AND ($4 = '' OR p.turno = $4)
		   AND ($5 = '' OR p.fecha >= $5::date)
		   AND ($6 = '' OR p.fecha <= $6::date)
		   AND (NOT $9 OR p.id_mina = ANY($10::uuid[]))
		   AND ($7 = '' OR (to_char(p.fecha, 'YYYY-MM-DD'), p.id::text) > ($7, $8))
		 ORDER BY p.fecha, p.id LIMIT $11`,
		filtro.Mina, filtro.Obra, filtro.Equipo, filtro.Turno, filtro.Desde, filtro.Hasta,
		orden, identificadorCursor, filtro.RestringirPorMina, filtro.MinasPermitidas, filtro.Limite+1)
	if err != nil {
		return nil, "", err
	}
	defer filas.Close()
	partes, err := leerResumenes(filas)
	if err != nil {
		return nil, "", err
	}
	siguiente := ""
	if len(partes) > filtro.Limite {
		partes = partes[:filtro.Limite]
		ultimo := partes[len(partes)-1]
		siguiente = paginacion.CodificarCursor(ultimo.Fecha, ultimo.Identificador)
	}
	return partes, siguiente, nil
}

func (LectorDeProduccionPostgres) DetalleDeParteDeBarrenacion(ctx context.Context, identificadorParte string) (puertos.DetalleDeParteDeBarrenacion, bool, error) {
	consultas := persistencia.ConsultasDe(ctx)
	var detalle puertos.DetalleDeParteDeBarrenacion
	fila := consultas.QueryRow(ctx,
		`SELECT p.id, to_char(p.fecha, 'YYYY-MM-DD'), p.turno, m.nombre, o.codigo,
		        COALESCE(eq.codigo, ''), op.nombre_completo,
		        COALESCE(to_char(p.horometro_percusion_final - p.horometro_percusion_inicial, 'FM999999990.00'), '0.00'),
		        COALESCE((SELECT to_char(sum(a.longitud), 'FM999999990.00') FROM produccion.barrenacion_avance a
		                  WHERE a.id_parte_barrenacion = p.id), '0.00'),
		        (SELECT count(*) FROM produccion.barrenacion_avance a WHERE a.id_parte_barrenacion = p.id),
		        p.tipo_barrenacion, cap.nombre_completo, COALESCE(p.observaciones, '')
		 FROM produccion.parte_barrenacion p
		 JOIN catalogos.mina m ON m.id = p.id_mina
		 JOIN catalogos.obra o ON o.id = p.id_obra
		 JOIN catalogos.equipo eq ON eq.id = p.id_equipo
		 JOIN catalogos.empleado op ON op.id = p.id_operador
		 JOIN catalogos.empleado cap ON cap.id = p.id_capitan_mina
		 WHERE p.id = $1 AND p.eliminado_en IS NULL`, identificadorParte)
	if err := fila.Scan(&detalle.Identificador, &detalle.Fecha, &detalle.Turno, &detalle.Mina, &detalle.Obra,
		&detalle.Equipo, &detalle.Operador, &detalle.Horas, &detalle.Toneladas, &detalle.Detalles,
		&detalle.TipoDeBarrenacion, &detalle.CapitanDeMina, &detalle.Observaciones); err != nil {
		if err.Error() == "no rows in result set" {
			return puertos.DetalleDeParteDeBarrenacion{}, false, nil
		}
		return puertos.DetalleDeParteDeBarrenacion{}, false, err
	}
	filas, err := consultas.Query(ctx,
		`SELECT a.id, a.actividad, a.lugar, COALESCE(a.no_secciones::text, ''),
		        to_char(a.longitud, 'FM999999990.00'), a.no_barrenos,
		        COALESCE((SELECT string_agg(tb.descripcion, ', ' ORDER BY tb.descripcion)
		                  FROM produccion.barrenacion_ejecutado be
		                  JOIN catalogos.tipo_barreno tb ON tb.id = be.id_tipo_barreno
		                  WHERE be.id_barrenacion_avance = a.id), '')
		 FROM produccion.barrenacion_avance a
		 WHERE a.id_parte_barrenacion = $1 ORDER BY a.creado_en`, identificadorParte)
	if err != nil {
		return puertos.DetalleDeParteDeBarrenacion{}, false, err
	}
	defer filas.Close()
	for filas.Next() {
		var linea puertos.LineaDeAvance
		var ejecutados string
		if err := filas.Scan(&linea.Identificador, &linea.Actividad, &linea.Lugar, &linea.NumeroDeSecciones,
			&linea.Longitud, &linea.NumeroDeBarrenos, &ejecutados); err != nil {
			return puertos.DetalleDeParteDeBarrenacion{}, false, err
		}
		if ejecutados != "" {
			linea.Ejecutados = append(linea.Ejecutados, ejecutados)
		}
		detalle.Avances = append(detalle.Avances, linea)
	}
	return detalle, true, filas.Err()
}

func (LectorDeProduccionPostgres) ListarDemoras(ctx context.Context, filtro puertos.FiltroDeProduccion) ([]puertos.ResumenDeDemora, string, error) {
	consultas := persistencia.ConsultasDe(ctx)
	orden, identificadorCursor, err := paginacion.DecodificarCursor(filtro.Cursor)
	if err != nil {
		return nil, "", err
	}
	filas, err := consultas.Query(ctx,
		`SELECT d.id, to_char(d.fecha, 'YYYY-MM-DD'), d.turno, m.nombre, COALESCE(eq.codigo, ''),
		        td.descripcion, d.minutos, COALESCE(d.observacion, '')
		 FROM produccion.demora_equipo d
		 JOIN catalogos.mina m ON m.id = d.id_mina
		 JOIN catalogos.equipo eq ON eq.id = d.id_equipo
		 JOIN catalogos.tipo_demora td ON td.id = d.id_tipo_demora
		 WHERE ($1 = '' OR d.id_mina::text = $1)
		   AND ($2 = '' OR d.id_equipo::text = $2)
		   AND ($3 = '' OR d.turno = $3)
		   AND ($4 = '' OR d.fecha >= $4::date)
		   AND ($5 = '' OR d.fecha <= $5::date)
		   AND (NOT $8 OR d.id_mina = ANY($9::uuid[]))
		   AND ($6 = '' OR (to_char(d.fecha, 'YYYY-MM-DD'), d.id::text) > ($6, $7))
		 ORDER BY d.fecha, d.id LIMIT $10`,
		filtro.Mina, filtro.Equipo, filtro.Turno, filtro.Desde, filtro.Hasta,
		orden, identificadorCursor, filtro.RestringirPorMina, filtro.MinasPermitidas, filtro.Limite+1)
	if err != nil {
		return nil, "", err
	}
	defer filas.Close()
	var demoras []puertos.ResumenDeDemora
	for filas.Next() {
		var demora puertos.ResumenDeDemora
		if err := filas.Scan(&demora.Identificador, &demora.Fecha, &demora.Turno, &demora.Mina,
			&demora.Equipo, &demora.TipoDeDemora, &demora.Minutos, &demora.Observacion); err != nil {
			return nil, "", err
		}
		demoras = append(demoras, demora)
	}
	if err := filas.Err(); err != nil {
		return nil, "", err
	}
	siguiente := ""
	if len(demoras) > filtro.Limite {
		demoras = demoras[:filtro.Limite]
		ultima := demoras[len(demoras)-1]
		siguiente = paginacion.CodificarCursor(ultima.Fecha, ultima.Identificador)
	}
	return demoras, siguiente, nil
}

func (LectorDeProduccionPostgres) Indicadores(ctx context.Context, filtro puertos.FiltroDeProduccion) (puertos.IndicadoresDeProduccion, error) {
	consultas := persistencia.ConsultasDe(ctx)
	var indicadores puertos.IndicadoresDeProduccion
	err := consultas.QueryRow(ctx,
		`WITH rango AS (SELECT $1::text AS mina, $2::text AS desde, $3::text AS hasta,
		                       $4::boolean AS restringir, $5::uuid[] AS permitidas)
		 SELECT
		   (SELECT count(*) FROM produccion.parte_acarreo p, rango r WHERE p.eliminado_en IS NULL
		      AND (r.mina = '' OR p.id_mina::text = r.mina)
		      AND (r.desde = '' OR p.fecha >= r.desde::date) AND (r.hasta = '' OR p.fecha <= r.hasta::date)
		      AND (NOT r.restringir OR p.id_mina = ANY(r.permitidas))),
		   (SELECT count(*) FROM produccion.parte_rezagado p, rango r WHERE p.eliminado_en IS NULL
		      AND (r.mina = '' OR p.id_mina::text = r.mina)
		      AND (r.desde = '' OR p.fecha >= r.desde::date) AND (r.hasta = '' OR p.fecha <= r.hasta::date)
		      AND (NOT r.restringir OR p.id_mina = ANY(r.permitidas))),
		   (SELECT count(*) FROM produccion.parte_barrenacion p, rango r WHERE p.eliminado_en IS NULL
		      AND (r.mina = '' OR p.id_mina::text = r.mina)
		      AND (r.desde = '' OR p.fecha >= r.desde::date) AND (r.hasta = '' OR p.fecha <= r.hasta::date)
		      AND (NOT r.restringir OR p.id_mina = ANY(r.permitidas))),
		   COALESCE((SELECT to_char(sum(v.toneladas), 'FM999999990.000') FROM produccion.acarreo_viaje v
		      JOIN produccion.parte_acarreo p ON p.id = v.id_parte_acarreo, rango r
		      WHERE p.eliminado_en IS NULL
		      AND (r.mina = '' OR p.id_mina::text = r.mina)
		      AND (r.desde = '' OR p.fecha >= r.desde::date) AND (r.hasta = '' OR p.fecha <= r.hasta::date)
		      AND (NOT r.restringir OR p.id_mina = ANY(r.permitidas))), '0.000'),
		   COALESCE((SELECT to_char(sum(c.toneladas), 'FM999999990.000') FROM produccion.rezagado_ciclo c
		      JOIN produccion.parte_rezagado p ON p.id = c.id_parte_rezagado, rango r
		      WHERE p.eliminado_en IS NULL
		      AND (r.mina = '' OR p.id_mina::text = r.mina)
		      AND (r.desde = '' OR p.fecha >= r.desde::date) AND (r.hasta = '' OR p.fecha <= r.hasta::date)
		      AND (NOT r.restringir OR p.id_mina = ANY(r.permitidas))), '0.000'),
		   COALESCE((SELECT to_char(sum(a.longitud), 'FM999999990.00') FROM produccion.barrenacion_avance a
		      JOIN produccion.parte_barrenacion p ON p.id = a.id_parte_barrenacion, rango r
		      WHERE p.eliminado_en IS NULL
		      AND (r.mina = '' OR p.id_mina::text = r.mina)
		      AND (r.desde = '' OR p.fecha >= r.desde::date) AND (r.hasta = '' OR p.fecha <= r.hasta::date)
		      AND (NOT r.restringir OR p.id_mina = ANY(r.permitidas))), '0.00'),
		   COALESCE((SELECT sum(d.minutos) FROM produccion.demora_equipo d, rango r
		      WHERE (r.mina = '' OR d.id_mina::text = r.mina)
		      AND (r.desde = '' OR d.fecha >= r.desde::date) AND (r.hasta = '' OR d.fecha <= r.hasta::date)
		      AND (NOT r.restringir OR d.id_mina = ANY(r.permitidas))), 0)`,
		filtro.Mina, filtro.Desde, filtro.Hasta, filtro.RestringirPorMina, filtro.MinasPermitidas).
		Scan(&indicadores.PartesDeAcarreo, &indicadores.PartesDeRezagado, &indicadores.PartesDeBarrenacion,
			&indicadores.ToneladasAcarreadas, &indicadores.ToneladasRezagadas,
			&indicadores.MetrosBarrenados, &indicadores.MinutosDeDemora)
	return indicadores, err
}

type filasConsultables interface {
	Next() bool
	Scan(destinos ...any) error
	Err() error
}

func leerResumenes(filas filasConsultables) ([]puertos.ResumenDeParte, error) {
	var partes []puertos.ResumenDeParte
	for filas.Next() {
		var parte puertos.ResumenDeParte
		if err := filas.Scan(&parte.Identificador, &parte.Fecha, &parte.Turno, &parte.Mina, &parte.Obra,
			&parte.Equipo, &parte.Operador, &parte.Horas, &parte.Toneladas, &parte.Detalles); err != nil {
			return nil, err
		}
		partes = append(partes, parte)
	}
	return partes, filas.Err()
}
