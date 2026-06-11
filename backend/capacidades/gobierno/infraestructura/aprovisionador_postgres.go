package infraestructura

import (
	"context"

	"minas/capacidades/gobierno/dominio"
	"minas/compartido/identificador"
	"minas/plataforma/persistencia"
)

type AprovisionadorDeAccesosPostgres struct{}

func NuevoAprovisionadorDeAccesos() AprovisionadorDeAccesosPostgres {
	return AprovisionadorDeAccesosPostgres{}
}

func (AprovisionadorDeAccesosPostgres) SembrarRolesDeSistema(ctx context.Context, idEmpresa identificador.Identificador, definiciones []dominio.DefinicionDeRolDeSistema) error {
	consultas := persistencia.ConsultasDe(ctx)
	for _, definicion := range definiciones {
		var idRol string
		err := consultas.QueryRow(ctx,
			`INSERT INTO gobierno.rol (id_empresa, codigo, descripcion, es_sistema)
			 VALUES ($1, $2, $3, true) RETURNING id`,
			idEmpresa.Texto(), definicion.Codigo, definicion.Descripcion).Scan(&idRol)
		if err != nil {
			return err
		}
		if len(definicion.Permisos) == 1 && definicion.Permisos[0] == dominio.TodosLosPermisos {
			if _, err := consultas.Exec(ctx,
				`INSERT INTO gobierno.rol_permiso (id_empresa, id_rol, id_permiso)
				 SELECT $1, $2, p.id FROM gobierno.permiso p WHERE p.eliminado_en IS NULL AND p.estado = 'ACTIVO'`,
				idEmpresa.Texto(), idRol); err != nil {
				return err
			}
			continue
		}
		if _, err := consultas.Exec(ctx,
			`INSERT INTO gobierno.rol_permiso (id_empresa, id_rol, id_permiso)
			 SELECT $1, $2, p.id FROM gobierno.permiso p WHERE p.codigo = ANY($3)`,
			idEmpresa.Texto(), idRol, definicion.Permisos); err != nil {
			return err
		}
	}
	return nil
}
