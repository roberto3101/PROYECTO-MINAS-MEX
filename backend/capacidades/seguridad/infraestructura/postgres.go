package infraestructura

import (
	"context"

	"minas/capacidades/seguridad/dominio"
	"minas/capacidades/seguridad/puertos"
	"minas/compartido/identificador"
	"minas/compartido/paginacion"
	"minas/plataforma/persistencia"
)

type RepositorioDeIncidentesPostgres struct{}

func NuevoRepositorioDeIncidentes() RepositorioDeIncidentesPostgres {
	return RepositorioDeIncidentesPostgres{}
}

func (RepositorioDeIncidentesPostgres) Guardar(ctx context.Context, incidente dominio.Incidente) error {
	consultas := persistencia.ConsultasDe(ctx)
	_, err := consultas.Exec(ctx,
		`INSERT INTO seguridad.incidente (id, id_empresa, id_mina, id_obra, id_tipo_incidente,
		                                  id_reportado_por, fecha, turno, severidad, descripcion, accion_inmediata)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NULLIF($11, ''))`,
		incidente.Identificador().Texto(), incidente.Empresa().Texto(), incidente.Mina().Texto(),
		textoOpcional(incidente.Obra()), incidente.TipoDeIncidente().Texto(),
		textoOpcional(incidente.ReportadoPor()), incidente.Fecha(), incidente.Turno(),
		incidente.Severidad(), incidente.Descripcion(), incidente.AccionInmediata())
	return err
}

type LectorDeSeguridadPostgres struct{}

func NuevoLectorDeSeguridad() LectorDeSeguridadPostgres {
	return LectorDeSeguridadPostgres{}
}

func (LectorDeSeguridadPostgres) ListarIncidentes(ctx context.Context, filtro puertos.FiltroDeIncidentes) ([]puertos.ResumenDeIncidente, string, error) {
	consultas := persistencia.ConsultasDe(ctx)
	orden, identificadorCursor, err := paginacion.DecodificarCursor(filtro.Cursor)
	if err != nil {
		return nil, "", err
	}
	filas, err := consultas.Query(ctx,
		`SELECT i.id, to_char(i.fecha, 'YYYY-MM-DD'), i.turno, m.nombre, COALESCE(o.codigo, ''),
		        t.descripcion, t.requiere_paro, i.severidad, i.descripcion,
		        COALESCE(i.accion_inmediata, ''), COALESCE(e.nombre_completo, '')
		 FROM seguridad.incidente i
		 JOIN catalogos.mina m ON m.id = i.id_mina
		 LEFT JOIN catalogos.obra o ON o.id = i.id_obra
		 JOIN seguridad.tipo_incidente t ON t.id = i.id_tipo_incidente
		 LEFT JOIN catalogos.empleado e ON e.id = i.id_reportado_por
		 WHERE ($1 = '' OR i.id_mina::text = $1)
		   AND ($2 = '' OR i.severidad = $2)
		   AND ($3 = '' OR i.fecha >= $3::date)
		   AND ($4 = '' OR i.fecha <= $4::date)
		   AND (NOT $7 OR i.id_mina = ANY($8::uuid[]))
		   AND ($5 = '' OR (to_char(i.fecha, 'YYYY-MM-DD'), i.id::text) > ($5, $6))
		 ORDER BY i.fecha, i.id LIMIT $9`,
		filtro.Mina, filtro.Severidad, filtro.Desde, filtro.Hasta,
		orden, identificadorCursor, filtro.RestringirPorMina, filtro.MinasPermitidas, filtro.Limite+1)
	if err != nil {
		return nil, "", err
	}
	defer filas.Close()
	var incidentes []puertos.ResumenDeIncidente
	for filas.Next() {
		var incidente puertos.ResumenDeIncidente
		if err := filas.Scan(&incidente.Identificador, &incidente.Fecha, &incidente.Turno, &incidente.Mina,
			&incidente.Obra, &incidente.TipoDeIncidente, &incidente.RequierePar, &incidente.Severidad,
			&incidente.Descripcion, &incidente.AccionInmediata, &incidente.ReportadoPor); err != nil {
			return nil, "", err
		}
		incidentes = append(incidentes, incidente)
	}
	if err := filas.Err(); err != nil {
		return nil, "", err
	}
	siguiente := ""
	if len(incidentes) > filtro.Limite {
		incidentes = incidentes[:filtro.Limite]
		ultimo := incidentes[len(incidentes)-1]
		siguiente = paginacion.CodificarCursor(ultimo.Fecha, ultimo.Identificador)
	}
	return incidentes, siguiente, nil
}

func (LectorDeSeguridadPostgres) ListarTiposDeIncidente(ctx context.Context) ([]puertos.OpcionDeSeguridad, error) {
	consultas := persistencia.ConsultasDe(ctx)
	filas, err := consultas.Query(ctx,
		`SELECT id, codigo, descripcion, requiere_paro FROM seguridad.tipo_incidente
		 WHERE eliminado_en IS NULL AND estado = 'ACTIVO' ORDER BY codigo`)
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	var tipos []puertos.OpcionDeSeguridad
	for filas.Next() {
		var tipo puertos.OpcionDeSeguridad
		if err := filas.Scan(&tipo.Identificador, &tipo.Codigo, &tipo.Descripcion, &tipo.RequiereParo); err != nil {
			return nil, err
		}
		tipos = append(tipos, tipo)
	}
	return tipos, filas.Err()
}

func (LectorDeSeguridadPostgres) ConteoPorSeveridad(ctx context.Context, filtro puertos.FiltroDeIncidentes) ([]puertos.ConteoPorSeveridad, error) {
	consultas := persistencia.ConsultasDe(ctx)
	filas, err := consultas.Query(ctx,
		`SELECT i.severidad, count(*)::int FROM seguridad.incidente i
		 WHERE ($1 = '' OR i.id_mina::text = $1)
		   AND ($2 = '' OR i.fecha >= $2::date)
		   AND ($3 = '' OR i.fecha <= $3::date)
		   AND (NOT $4 OR i.id_mina = ANY($5::uuid[]))
		 GROUP BY i.severidad ORDER BY i.severidad`,
		filtro.Mina, filtro.Desde, filtro.Hasta, filtro.RestringirPorMina, filtro.MinasPermitidas)
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	var conteo []puertos.ConteoPorSeveridad
	for filas.Next() {
		var linea puertos.ConteoPorSeveridad
		if err := filas.Scan(&linea.Severidad, &linea.Total); err != nil {
			return nil, err
		}
		conteo = append(conteo, linea)
	}
	return conteo, filas.Err()
}

func textoOpcional(valor *identificador.Identificador) any {
	if valor == nil {
		return nil
	}
	return valor.Texto()
}
