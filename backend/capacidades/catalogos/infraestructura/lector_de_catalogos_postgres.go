package infraestructura

import (
	"context"

	"minas/capacidades/catalogos/puertos"
	"minas/compartido/paginacion"
	"minas/plataforma/persistencia"
)

type LectorDeCatalogosPostgres struct{}

func NuevoLectorDeCatalogos() LectorDeCatalogosPostgres {
	return LectorDeCatalogosPostgres{}
}

func (LectorDeCatalogosPostgres) ListarMinas(ctx context.Context, filtro puertos.FiltroDeCatalogo) ([]puertos.ResumenMina, string, error) {
	consultas := persistencia.ConsultasDe(ctx)
	orden, identificadorCursor, err := paginacion.DecodificarCursor(filtro.Cursor)
	if err != nil {
		return nil, "", err
	}
	filas, err := consultas.Query(ctx,
		`SELECT id, nombre, area, COALESCE(niveles, ''), estado
		 FROM catalogos.mina
		 WHERE eliminado_en IS NULL
		   AND ($1 = '' OR nombre ILIKE '%'||$1||'%' OR area ILIKE '%'||$1||'%')
		   AND ($2 = '' OR $2 IN ('TODAS', 'TODOS') OR estado = $2)
		   AND (NOT $6 OR id = ANY($7::uuid[]))
		   AND ($3 = '' OR (nombre, id::text) > ($3, $4))
		 ORDER BY nombre, id LIMIT $5`,
		filtro.Busqueda, filtro.Estado, orden, identificadorCursor, filtro.Limite+1,
		filtro.RestringirPorMina, filtro.MinasPermitidas)
	if err != nil {
		return nil, "", err
	}
	defer filas.Close()
	var minas []puertos.ResumenMina
	for filas.Next() {
		var mina puertos.ResumenMina
		if err := filas.Scan(&mina.Identificador, &mina.Nombre, &mina.Area, &mina.Niveles, &mina.Estado); err != nil {
			return nil, "", err
		}
		minas = append(minas, mina)
	}
	if err := filas.Err(); err != nil {
		return nil, "", err
	}
	siguiente := ""
	if len(minas) > filtro.Limite {
		minas = minas[:filtro.Limite]
		ultima := minas[len(minas)-1]
		siguiente = paginacion.CodificarCursor(ultima.Nombre, ultima.Identificador)
	}
	return minas, siguiente, nil
}

func (LectorDeCatalogosPostgres) DetalleDeMina(ctx context.Context, identificadorMina string) (puertos.DetalleMina, bool, error) {
	consultas := persistencia.ConsultasDe(ctx)
	var detalle puertos.DetalleMina
	fila := consultas.QueryRow(ctx,
		`SELECT id, nombre, area, COALESCE(niveles, ''), estado,
		        COALESCE(densidad_promedio::text, ''), COALESCE(seccion_obra_largo::text, ''),
		        COALESCE(seccion_obra_ancho::text, ''), COALESCE(seccion_veta_largo::text, ''),
		        COALESCE(seccion_veta_ancho::text, ''),
		        to_char(creado_en, 'YYYY-MM-DD"T"HH24:MI:SSZ')
		 FROM catalogos.mina WHERE id = $1 AND eliminado_en IS NULL`, identificadorMina)
	err := fila.Scan(&detalle.Identificador, &detalle.Nombre, &detalle.Area, &detalle.Niveles, &detalle.Estado,
		&detalle.DensidadPromedio, &detalle.SeccionObraLargo, &detalle.SeccionObraAncho,
		&detalle.SeccionVetaLargo, &detalle.SeccionVetaAncho, &detalle.CreadoEn)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return puertos.DetalleMina{}, false, nil
		}
		return puertos.DetalleMina{}, false, err
	}
	return detalle, true, nil
}

func (LectorDeCatalogosPostgres) ListarEmpleados(ctx context.Context, filtro puertos.FiltroDeCatalogo) ([]puertos.ResumenEmpleado, string, error) {
	consultas := persistencia.ConsultasDe(ctx)
	orden, identificadorCursor, err := paginacion.DecodificarCursor(filtro.Cursor)
	if err != nil {
		return nil, "", err
	}
	filas, err := consultas.Query(ctx,
		`SELECT e.id, e.numero_nomina, e.nombre_completo, m.nombre,
		        COALESCE(d.descripcion, ''), COALESCE(pu.descripcion, ''), COALESCE(e.grupo, ''), e.estado
		 FROM catalogos.empleado e
		 JOIN catalogos.mina m ON m.id = e.id_mina
		 LEFT JOIN catalogos.departamento d ON d.id = e.id_departamento
		 LEFT JOIN catalogos.puesto pu ON pu.id = e.id_puesto
		 WHERE e.eliminado_en IS NULL
		   AND ($1 = '' OR e.numero_nomina ILIKE '%'||$1||'%' OR e.nombre_completo ILIKE '%'||$1||'%')
		   AND ($2 = '' OR $2 IN ('TODAS', 'TODOS') OR e.estado = $2)
		   AND (NOT $6 OR e.id_mina = ANY($7::uuid[]))
		   AND ($3 = '' OR (e.nombre_completo, e.id::text) > ($3, $4))
		 ORDER BY e.nombre_completo, e.id LIMIT $5`,
		filtro.Busqueda, filtro.Estado, orden, identificadorCursor, filtro.Limite+1,
		filtro.RestringirPorMina, filtro.MinasPermitidas)
	if err != nil {
		return nil, "", err
	}
	defer filas.Close()
	var empleados []puertos.ResumenEmpleado
	for filas.Next() {
		var empleado puertos.ResumenEmpleado
		if err := filas.Scan(&empleado.Identificador, &empleado.NumeroNomina, &empleado.NombreCompleto,
			&empleado.Mina, &empleado.Departamento, &empleado.Puesto, &empleado.Grupo, &empleado.Estado); err != nil {
			return nil, "", err
		}
		empleados = append(empleados, empleado)
	}
	if err := filas.Err(); err != nil {
		return nil, "", err
	}
	siguiente := ""
	if len(empleados) > filtro.Limite {
		empleados = empleados[:filtro.Limite]
		ultimo := empleados[len(empleados)-1]
		siguiente = paginacion.CodificarCursor(ultimo.NombreCompleto, ultimo.Identificador)
	}
	return empleados, siguiente, nil
}

func (LectorDeCatalogosPostgres) DetalleDeEmpleado(ctx context.Context, identificadorEmpleado string) (puertos.DetalleEmpleado, bool, error) {
	consultas := persistencia.ConsultasDe(ctx)
	var detalle puertos.DetalleEmpleado
	fila := consultas.QueryRow(ctx,
		`SELECT e.id, e.numero_nomina, e.nombre_completo, m.nombre,
		        COALESCE(d.descripcion, ''), COALESCE(pu.descripcion, ''), COALESCE(e.grupo, ''), e.estado,
		        COALESCE(a.descripcion, ''), COALESCE(e.centro_costo, ''), COALESCE(e.gerente_a_cargo, ''),
		        to_char(e.creado_en, 'YYYY-MM-DD"T"HH24:MI:SSZ')
		 FROM catalogos.empleado e
		 JOIN catalogos.mina m ON m.id = e.id_mina
		 LEFT JOIN catalogos.departamento d ON d.id = e.id_departamento
		 LEFT JOIN catalogos.puesto pu ON pu.id = e.id_puesto
		 LEFT JOIN catalogos.actividad a ON a.id = e.id_actividad
		 WHERE e.id = $1 AND e.eliminado_en IS NULL`, identificadorEmpleado)
	err := fila.Scan(&detalle.Identificador, &detalle.NumeroNomina, &detalle.NombreCompleto, &detalle.Mina,
		&detalle.Departamento, &detalle.Puesto, &detalle.Grupo, &detalle.Estado,
		&detalle.Actividad, &detalle.CentroCosto, &detalle.GerenteACargo, &detalle.CreadoEn)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return puertos.DetalleEmpleado{}, false, nil
		}
		return puertos.DetalleEmpleado{}, false, err
	}
	return detalle, true, nil
}

func (LectorDeCatalogosPostgres) ListarEquipos(ctx context.Context, filtro puertos.FiltroDeCatalogo) ([]puertos.ResumenEquipo, string, error) {
	consultas := persistencia.ConsultasDe(ctx)
	orden, identificadorCursor, err := paginacion.DecodificarCursor(filtro.Cursor)
	if err != nil {
		return nil, "", err
	}
	filas, err := consultas.Query(ctx,
		`SELECT e.id, COALESCE(e.codigo, ''), t.descripcion, mt.descripcion, m.nombre,
		        COALESCE(e.fabricante, ''), COALESCE(e.modelo, ''), e.estado
		 FROM catalogos.equipo e
		 JOIN catalogos.mina m ON m.id = e.id_mina
		 JOIN catalogos.tipo_equipo t ON t.id = e.id_tipo_equipo
		 JOIN catalogos.modulo_trabajo mt ON mt.id = e.id_modulo_trabajo
		 WHERE e.eliminado_en IS NULL
		   AND ($1 = '' OR e.codigo ILIKE '%'||$1||'%' OR COALESCE(e.fabricante, '') ILIKE '%'||$1||'%' OR COALESCE(e.modelo, '') ILIKE '%'||$1||'%')
		   AND ($2 = '' OR $2 IN ('TODAS', 'TODOS') OR e.estado = $2)
		   AND (NOT $6 OR e.id_mina = ANY($7::uuid[]))
		   AND ($3 = '' OR (e.codigo, e.id::text) > ($3, $4))
		 ORDER BY e.codigo, e.id LIMIT $5`,
		filtro.Busqueda, filtro.Estado, orden, identificadorCursor, filtro.Limite+1,
		filtro.RestringirPorMina, filtro.MinasPermitidas)
	if err != nil {
		return nil, "", err
	}
	defer filas.Close()
	var equipos []puertos.ResumenEquipo
	for filas.Next() {
		var equipo puertos.ResumenEquipo
		if err := filas.Scan(&equipo.Identificador, &equipo.Codigo, &equipo.Tipo, &equipo.Modulo,
			&equipo.Mina, &equipo.Fabricante, &equipo.Modelo, &equipo.Estado); err != nil {
			return nil, "", err
		}
		equipos = append(equipos, equipo)
	}
	if err := filas.Err(); err != nil {
		return nil, "", err
	}
	siguiente := ""
	if len(equipos) > filtro.Limite {
		equipos = equipos[:filtro.Limite]
		ultimo := equipos[len(equipos)-1]
		siguiente = paginacion.CodificarCursor(ultimo.Codigo, ultimo.Identificador)
	}
	return equipos, siguiente, nil
}

func (LectorDeCatalogosPostgres) DetalleDeEquipo(ctx context.Context, identificadorEquipo string) (puertos.DetalleEquipo, bool, error) {
	consultas := persistencia.ConsultasDe(ctx)
	var detalle puertos.DetalleEquipo
	fila := consultas.QueryRow(ctx,
		`SELECT e.id, COALESCE(e.codigo, ''), t.descripcion, mt.descripcion, m.nombre,
		        COALESCE(e.fabricante, ''), COALESCE(e.modelo, ''), e.estado,
		        COALESCE(e.descripcion, ''), COALESCE(e.numero_serie, ''),
		        COALESCE(to_char(e.fecha_ingreso_mina, 'YYYY-MM-DD'), ''),
		        COALESCE(e.modelo_perforadora, ''), COALESCE(e.capacidad_longitud::text, ''),
		        COALESCE(e.anio_fabricacion, 0),
		        to_char(e.creado_en, 'YYYY-MM-DD"T"HH24:MI:SSZ')
		 FROM catalogos.equipo e
		 JOIN catalogos.mina m ON m.id = e.id_mina
		 JOIN catalogos.tipo_equipo t ON t.id = e.id_tipo_equipo
		 JOIN catalogos.modulo_trabajo mt ON mt.id = e.id_modulo_trabajo
		 WHERE e.id = $1 AND e.eliminado_en IS NULL`, identificadorEquipo)
	err := fila.Scan(&detalle.Identificador, &detalle.Codigo, &detalle.Tipo, &detalle.Modulo, &detalle.Mina,
		&detalle.Fabricante, &detalle.Modelo, &detalle.Estado,
		&detalle.Descripcion, &detalle.NumeroSerie, &detalle.FechaIngresoMina,
		&detalle.ModeloPerforadora, &detalle.CapacidadLongitud, &detalle.AnioFabricacion, &detalle.CreadoEn)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return puertos.DetalleEquipo{}, false, nil
		}
		return puertos.DetalleEquipo{}, false, err
	}
	return detalle, true, nil
}

func (LectorDeCatalogosPostgres) ListarTiposDeEquipo(ctx context.Context) ([]puertos.OpcionDeCatalogo, error) {
	return listarOpciones(ctx, "catalogos.tipo_equipo")
}

func (LectorDeCatalogosPostgres) ListarModulosDeTrabajo(ctx context.Context) ([]puertos.OpcionDeCatalogo, error) {
	return listarOpciones(ctx, "catalogos.modulo_trabajo")
}

func (LectorDeCatalogosPostgres) ListarDepartamentos(ctx context.Context) ([]puertos.OpcionDeCatalogo, error) {
	return listarOpciones(ctx, "catalogos.departamento")
}

func (LectorDeCatalogosPostgres) ListarPuestos(ctx context.Context) ([]puertos.OpcionDeCatalogo, error) {
	return listarOpciones(ctx, "catalogos.puesto")
}

func (LectorDeCatalogosPostgres) ListarActividades(ctx context.Context) ([]puertos.OpcionDeCatalogo, error) {
	return listarOpciones(ctx, "catalogos.actividad")
}

func listarOpciones(ctx context.Context, tabla string) ([]puertos.OpcionDeCatalogo, error) {
	consultas := persistencia.ConsultasDe(ctx)
	filas, err := consultas.Query(ctx,
		`SELECT id, codigo, descripcion FROM `+tabla+` WHERE eliminado_en IS NULL ORDER BY codigo`)
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	var opciones []puertos.OpcionDeCatalogo
	for filas.Next() {
		var opcion puertos.OpcionDeCatalogo
		if err := filas.Scan(&opcion.Identificador, &opcion.Codigo, &opcion.Descripcion); err != nil {
			return nil, err
		}
		opciones = append(opciones, opcion)
	}
	return opciones, filas.Err()
}
