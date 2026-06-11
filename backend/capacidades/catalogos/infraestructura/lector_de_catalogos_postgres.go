package infraestructura

import (
	"context"

	"minas/capacidades/catalogos/puertos"
	"minas/plataforma/persistencia"
)

type LectorDeCatalogosPostgres struct{}

func NuevoLectorDeCatalogos() LectorDeCatalogosPostgres {
	return LectorDeCatalogosPostgres{}
}

func (LectorDeCatalogosPostgres) ListarMinas(ctx context.Context) ([]puertos.ResumenMina, error) {
	consultas := persistencia.ConsultasDe(ctx)
	filas, err := consultas.Query(ctx,
		`SELECT id, nombre, COALESCE(area, '')
		 FROM catalogos.mina WHERE eliminado_en IS NULL ORDER BY nombre`)
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	var minas []puertos.ResumenMina
	for filas.Next() {
		var mina puertos.ResumenMina
		if err := filas.Scan(&mina.Identificador, &mina.Nombre, &mina.Area); err != nil {
			return nil, err
		}
		minas = append(minas, mina)
	}
	return minas, filas.Err()
}

func (LectorDeCatalogosPostgres) ListarEmpleados(ctx context.Context) ([]puertos.ResumenEmpleado, error) {
	consultas := persistencia.ConsultasDe(ctx)
	filas, err := consultas.Query(ctx,
		`SELECT e.id, e.numero_nomina, e.nombre_completo, m.nombre
		 FROM catalogos.empleado e
		 JOIN catalogos.mina m ON m.id = e.id_mina
		 WHERE e.eliminado_en IS NULL ORDER BY e.nombre_completo`)
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	var empleados []puertos.ResumenEmpleado
	for filas.Next() {
		var empleado puertos.ResumenEmpleado
		if err := filas.Scan(&empleado.Identificador, &empleado.NumeroNomina, &empleado.NombreCompleto, &empleado.Mina); err != nil {
			return nil, err
		}
		empleados = append(empleados, empleado)
	}
	return empleados, filas.Err()
}

func (LectorDeCatalogosPostgres) ListarEquipos(ctx context.Context) ([]puertos.ResumenEquipo, error) {
	consultas := persistencia.ConsultasDe(ctx)
	filas, err := consultas.Query(ctx,
		`SELECT e.id, e.codigo, m.nombre, t.descripcion, mt.descripcion, COALESCE(e.fabricante, '')
		 FROM catalogos.equipo e
		 JOIN catalogos.mina m ON m.id = e.id_mina
		 JOIN catalogos.tipo_equipo t ON t.id = e.id_tipo_equipo
		 JOIN catalogos.modulo_trabajo mt ON mt.id = e.id_modulo_trabajo
		 WHERE e.eliminado_en IS NULL ORDER BY e.codigo`)
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	var equipos []puertos.ResumenEquipo
	for filas.Next() {
		var equipo puertos.ResumenEquipo
		if err := filas.Scan(&equipo.Identificador, &equipo.Codigo, &equipo.Mina, &equipo.Tipo, &equipo.Modulo, &equipo.Fabricante); err != nil {
			return nil, err
		}
		equipos = append(equipos, equipo)
	}
	return equipos, filas.Err()
}

func (LectorDeCatalogosPostgres) ListarTiposDeEquipo(ctx context.Context) ([]puertos.OpcionDeCatalogo, error) {
	return listarOpciones(ctx, "catalogos.tipo_equipo")
}

func (LectorDeCatalogosPostgres) ListarModulosDeTrabajo(ctx context.Context) ([]puertos.OpcionDeCatalogo, error) {
	return listarOpciones(ctx, "catalogos.modulo_trabajo")
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
