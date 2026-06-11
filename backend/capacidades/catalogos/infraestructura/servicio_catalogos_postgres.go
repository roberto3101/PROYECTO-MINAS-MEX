package infraestructura

import (
	"context"

	"minas/capacidades/catalogos/contrato"
	"minas/capacidades/catalogos/puertos"
	"minas/plataforma/persistencia"
)

type ServicioCatalogosPostgres struct {
	unidad puertos.UnidadDeTrabajo
}

func NuevoServicioCatalogos(unidad puertos.UnidadDeTrabajo) ServicioCatalogosPostgres {
	return ServicioCatalogosPostgres{unidad: unidad}
}

func (servicio ServicioCatalogosPostgres) MinasActivas(ctx context.Context) ([]contrato.MinaPublicada, error) {
	var minas []contrato.MinaPublicada
	err := servicio.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		consultas := persistencia.ConsultasDe(ctx)
		filas, err := consultas.Query(ctx,
			`SELECT id, nombre FROM catalogos.mina
			 WHERE eliminado_en IS NULL AND estado = 'ACTIVA' ORDER BY nombre`)
		if err != nil {
			return err
		}
		defer filas.Close()
		for filas.Next() {
			var mina contrato.MinaPublicada
			if err := filas.Scan(&mina.Identificador, &mina.Nombre); err != nil {
				return err
			}
			minas = append(minas, mina)
		}
		return filas.Err()
	})
	return minas, err
}

func (ServicioCatalogosPostgres) SembrarCatalogosBasicos(ctx context.Context, identificadorEmpresa string) error {
	consultas := persistencia.ConsultasDe(ctx)
	var sembrado bool
	if err := consultas.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM catalogos.tipo_obra WHERE id_empresa = $1)`,
		identificadorEmpresa).Scan(&sembrado); err != nil {
		return err
	}
	if sembrado {
		return nil
	}
	instrucciones := []string{
		`INSERT INTO catalogos.tipo_obra (id_empresa, codigo, descripcion) VALUES
		 ($1, 'RPA', 'Rampa'), ($1, 'LAT', 'Lateral'), ($1, 'ACC', 'Acceso'), ($1, 'CRO', 'Crucero'),
		 ($1, 'REB', 'Rebaje'), ($1, 'STOCK', 'Stock'), ($1, 'SILL', 'Sill (galeria de base)')`,
		`INSERT INTO catalogos.sistema_minado (id_empresa, codigo, descripcion) VALUES
		 ($1, 'BL', 'Bench and Long'), ($1, 'CR', 'Corte y relleno'), ($1, 'TSC', 'Tumbe sobre carga')`,
		`INSERT INTO catalogos.tipo_equipo (id_empresa, codigo, descripcion) VALUES
		 ($1, 'CAMION', 'Camion'), ($1, 'SCOOP', 'Scooptram'), ($1, 'JUMBO_LINEAL', 'Jumbo lineal'),
		 ($1, 'JUMBO_LARGA', 'Jumbo barrenacion larga'), ($1, 'ANCLADOR', 'Anclador'),
		 ($1, 'LOCOMOTORA', 'Locomotora'), ($1, 'TRACTOR', 'Tractor'),
		 ($1, 'TIRO', 'Tiro / Shaft (izaje vertical a superficie)'), ($1, 'BANDA', 'Banda transportadora de extraccion')`,
		`INSERT INTO catalogos.modulo_trabajo (id_empresa, codigo, descripcion) VALUES
		 ($1, 'ACARREO', 'Acarreo'), ($1, 'REZAGADO', 'Rezagado'), ($1, 'BARRENACION', 'Barrenacion'),
		 ($1, 'ANCLAJE', 'Anclaje'), ($1, 'RELLENO', 'Relleno')`,
		`INSERT INTO catalogos.tipo_mineral (id_empresa, codigo, descripcion, es_mineral) VALUES
		 ($1, 'MINERAL', 'Mineral', true), ($1, 'TEPETATE', 'Tepetate', false)`,
		`INSERT INTO catalogos.departamento (id_empresa, codigo, descripcion) VALUES
		 ($1, 'MINA', 'Mina'), ($1, 'PLANEACION', 'Planeacion'), ($1, 'GERENCIA', 'Gerencia')`,
		`INSERT INTO catalogos.puesto (id_empresa, codigo, descripcion) VALUES
		 ($1, 'CAP_MINA', 'Capitan de Mina'), ($1, 'SUPERVISOR', 'Supervisor'), ($1, 'OPERADOR', 'Operador')`,
		`INSERT INTO catalogos.actividad (id_empresa, codigo, descripcion) VALUES
		 ($1, 'DESARROLLO', 'Desarrollo'), ($1, 'TUMBE', 'Tumbe'), ($1, 'BARRENACION', 'Barrenacion'),
		 ($1, 'REZAGADO', 'Rezagado'), ($1, 'ACARREO', 'Acarreo'), ($1, 'ANCLAJE', 'Anclaje')`,
		`INSERT INTO catalogos.tipo_demora (id_empresa, codigo, descripcion, es_programada) VALUES
		 ($1, 'MECANICA', 'Falla mecanica', false), ($1, 'ELECTRICA', 'Falla electrica', false),
		 ($1, 'OPERATIVA', 'Demora operativa', false), ($1, 'PROGRAMADA', 'Mantenimiento programado', true)`,
		`INSERT INTO catalogos.tipo_explosivo (id_empresa, codigo, descripcion, familia, unidad_medida) VALUES
		 ($1, 'ANFO', 'Anfo', 'AGENTE', 'kg'), ($1, 'EMULSION', 'Emulsion', 'AGENTE', 'kg'),
		 ($1, 'FULMINANTE', 'Fulminante', 'INICIADOR', 'pza'), ($1, 'CORDON', 'Cordon detonante', 'CORDON', 'm')`,
		`INSERT INTO catalogos.tipo_barreno (id_empresa, codigo, descripcion) VALUES
		 ($1, 'CUELE', 'Cuele (arranque)'), ($1, 'CONTRACUELE', 'Contracuele'), ($1, 'DESTROZA', 'Destroza / ensanche'),
		 ($1, 'CONTORNO', 'Contorno / recorte'), ($1, 'ZAPATERA', 'Zapatera / piso (arrastre)'),
		 ($1, 'CORONA', 'Corona / hastiales'), ($1, 'AUXILIAR', 'Auxiliar')`,
	}
	for _, instruccion := range instrucciones {
		if _, err := consultas.Exec(ctx, instruccion, identificadorEmpresa); err != nil {
			return err
		}
	}
	return nil
}
