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

func (AprovisionadorDeAccesosPostgres) SembrarRolesDeSistema(ctx context.Context, idEmpresa identificador.Identificador, definiciones []dominio.DefinicionDeRolDeSistema) (map[string]identificador.Identificador, error) {
	consultas := persistencia.ConsultasDe(ctx)
	rolesSembrados := make(map[string]identificador.Identificador, len(definiciones))
	for _, definicion := range definiciones {
		var crudoIdRol string
		err := consultas.QueryRow(ctx,
			`INSERT INTO gobierno.rol (id_empresa, codigo, descripcion, es_sistema)
			 VALUES ($1, $2, $3, true) RETURNING id`,
			idEmpresa.Texto(), definicion.Codigo, definicion.Descripcion).Scan(&crudoIdRol)
		if err != nil {
			return nil, err
		}
		idRol, err := identificador.Desde(crudoIdRol)
		if err != nil {
			return nil, err
		}
		rolesSembrados[definicion.Codigo] = idRol
		if len(definicion.Permisos) == 1 && definicion.Permisos[0] == dominio.TodosLosPermisos {
			if _, err := consultas.Exec(ctx,
				`INSERT INTO gobierno.rol_permiso (id_empresa, id_rol, id_permiso)
				 SELECT $1, $2, p.id FROM gobierno.permiso p WHERE p.eliminado_en IS NULL AND p.estado = 'ACTIVO'`,
				idEmpresa.Texto(), crudoIdRol); err != nil {
				return nil, err
			}
			continue
		}
		if _, err := consultas.Exec(ctx,
			`INSERT INTO gobierno.rol_permiso (id_empresa, id_rol, id_permiso)
			 SELECT $1, $2, p.id FROM gobierno.permiso p WHERE p.codigo = ANY($3)`,
			idEmpresa.Texto(), crudoIdRol, definicion.Permisos); err != nil {
			return nil, err
		}
	}
	return rolesSembrados, nil
}
