package infraestructura

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"minas/capacidades/gobierno/dominio"
	"minas/compartido/identificador"
	"minas/plataforma/persistencia"
)

type RepositorioRolPostgres struct{}

func NuevoRepositorioRol() RepositorioRolPostgres {
	return RepositorioRolPostgres{}
}

func (RepositorioRolPostgres) Guardar(ctx context.Context, rol dominio.Rol) error {
	consultas := persistencia.ConsultasDe(ctx)
	if _, err := consultas.Exec(ctx,
		`INSERT INTO gobierno.rol (id, id_empresa, codigo, descripcion, es_sistema, estado)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (id) DO UPDATE SET descripcion = EXCLUDED.descripcion, estado = EXCLUDED.estado, actualizado_en = now()`,
		rol.Identificador().Texto(), rol.Empresa().Texto(), rol.Codigo(), rol.Descripcion(), rol.EsDeSistema(), string(rol.Estado())); err != nil {
		return err
	}
	permisosDeseados := textos(rol.Permisos())
	if _, err := consultas.Exec(ctx,
		`UPDATE gobierno.rol_permiso SET eliminado_en = now()
		 WHERE id_rol = $1 AND eliminado_en IS NULL AND id_permiso <> ALL ($2::uuid[])`,
		rol.Identificador().Texto(), permisosDeseados); err != nil {
		return err
	}
	_, err := consultas.Exec(ctx,
		`INSERT INTO gobierno.rol_permiso (id_empresa, id_rol, id_permiso)
		 SELECT $1, $2, permiso FROM unnest($3::uuid[]) AS permiso
		 ON CONFLICT (id_empresa, id_rol, id_permiso) WHERE eliminado_en IS NULL DO NOTHING`,
		rol.Empresa().Texto(), rol.Identificador().Texto(), permisosDeseados)
	return err
}

func (RepositorioRolPostgres) BuscarPorIdentificador(ctx context.Context, id identificador.Identificador) (dominio.Rol, bool, error) {
	return leerRol(ctx, "r.id = $1", id.Texto())
}

func (RepositorioRolPostgres) BuscarPorCodigo(ctx context.Context, codigo string) (dominio.Rol, bool, error) {
	return leerRol(ctx, "r.codigo = $1", codigo)
}

func leerRol(ctx context.Context, filtro, argumento string) (dominio.Rol, bool, error) {
	consultas := persistencia.ConsultasDe(ctx)
	var id, idEmpresa, codigo, descripcion, estado string
	var esDeSistema bool
	var permisos []string
	fila := consultas.QueryRow(ctx,
		`SELECT r.id, r.id_empresa, r.codigo, r.descripcion, r.es_sistema, r.estado,
		        COALESCE(array_agg(rp.id_permiso::text) FILTER (WHERE rp.id IS NOT NULL), '{}')
		 FROM gobierno.rol r
		 LEFT JOIN gobierno.rol_permiso rp ON rp.id_rol = r.id AND rp.eliminado_en IS NULL
		 WHERE `+filtro+` AND r.eliminado_en IS NULL
		 GROUP BY r.id`,
		argumento)
	if err := fila.Scan(&id, &idEmpresa, &codigo, &descripcion, &esDeSistema, &estado, &permisos); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dominio.Rol{}, false, nil
		}
		return dominio.Rol{}, false, err
	}
	identificadorRol, err := identificador.Desde(id)
	if err != nil {
		return dominio.Rol{}, false, err
	}
	identificadorEmpresa, err := identificador.Desde(idEmpresa)
	if err != nil {
		return dominio.Rol{}, false, err
	}
	permisosConcedidos, err := identificadores(permisos)
	if err != nil {
		return dominio.Rol{}, false, err
	}
	return dominio.ReconstruirRol(identificadorRol, identificadorEmpresa, codigo, descripcion, esDeSistema, permisosConcedidos, dominio.EstadoRol(estado)), true, nil
}
