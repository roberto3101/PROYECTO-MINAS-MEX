package infraestructura

import (
	"context"

	"minas/capacidades/gobierno/puertos"
	"minas/plataforma/persistencia"
)

type LectorDeGobiernoPostgres struct{}

func NuevoLectorDeGobierno() LectorDeGobiernoPostgres {
	return LectorDeGobiernoPostgres{}
}

func (LectorDeGobiernoPostgres) ListarUsuarios(ctx context.Context) ([]puertos.ResumenUsuario, error) {
	consultas := persistencia.ConsultasDe(ctx)
	filas, err := consultas.Query(ctx,
		`SELECT id, usuario, nombre, COALESCE(correo, ''), estado
		 FROM gobierno.usuario WHERE eliminado_en IS NULL ORDER BY usuario`)
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	var usuarios []puertos.ResumenUsuario
	for filas.Next() {
		var usuario puertos.ResumenUsuario
		if err := filas.Scan(&usuario.Identificador, &usuario.NombreCorto, &usuario.Nombre, &usuario.Correo, &usuario.Estado); err != nil {
			return nil, err
		}
		usuarios = append(usuarios, usuario)
	}
	return usuarios, filas.Err()
}

func (LectorDeGobiernoPostgres) ListarRoles(ctx context.Context) ([]puertos.ResumenRol, error) {
	consultas := persistencia.ConsultasDe(ctx)
	filas, err := consultas.Query(ctx,
		`SELECT r.id, r.codigo, r.descripcion, r.es_sistema,
		        COALESCE(array_agg(p.codigo ORDER BY p.codigo) FILTER (WHERE p.id IS NOT NULL), '{}')
		 FROM gobierno.rol r
		 LEFT JOIN gobierno.rol_permiso rp ON rp.id_rol = r.id AND rp.eliminado_en IS NULL
		 LEFT JOIN gobierno.permiso p ON p.id = rp.id_permiso AND p.eliminado_en IS NULL
		 WHERE r.eliminado_en IS NULL AND r.estado = 'ACTIVO'
		 GROUP BY r.id ORDER BY r.es_sistema DESC, r.codigo`)
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	var roles []puertos.ResumenRol
	for filas.Next() {
		var rol puertos.ResumenRol
		if err := filas.Scan(&rol.Identificador, &rol.Codigo, &rol.Descripcion, &rol.EsDeSistema, &rol.Permisos); err != nil {
			return nil, err
		}
		roles = append(roles, rol)
	}
	return roles, filas.Err()
}

func (LectorDeGobiernoPostgres) CatalogoDePermisos(ctx context.Context) ([]puertos.ResumenPermiso, error) {
	consultas := persistencia.ConsultasDe(ctx)
	filas, err := consultas.Query(ctx,
		`SELECT id, codigo, descripcion, modulo FROM gobierno.permiso
		 WHERE eliminado_en IS NULL AND estado = 'ACTIVO' ORDER BY modulo, codigo`)
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	var permisos []puertos.ResumenPermiso
	for filas.Next() {
		var permiso puertos.ResumenPermiso
		if err := filas.Scan(&permiso.Identificador, &permiso.Codigo, &permiso.Descripcion, &permiso.Modulo); err != nil {
			return nil, err
		}
		permisos = append(permisos, permiso)
	}
	return permisos, filas.Err()
}

func (LectorDeGobiernoPostgres) AsignacionesDe(ctx context.Context, identificadorUsuario string) ([]puertos.ResumenAsignacion, error) {
	consultas := persistencia.ConsultasDe(ctx)
	filas, err := consultas.Query(ctx,
		`SELECT ur.id, r.descripcion, r.codigo, COALESCE(ur.id_mina::text, ''), (ur.eliminado_en IS NULL)
		 FROM gobierno.usuario_rol ur
		 JOIN gobierno.rol r ON r.id = ur.id_rol
		 WHERE ur.id_usuario = $1
		 ORDER BY (ur.eliminado_en IS NULL) DESC, r.codigo`,
		identificadorUsuario)
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	var asignaciones []puertos.ResumenAsignacion
	for filas.Next() {
		var asignacion puertos.ResumenAsignacion
		if err := filas.Scan(&asignacion.Identificador, &asignacion.Rol, &asignacion.CodigoRol, &asignacion.AlcanceMina, &asignacion.Vigente); err != nil {
			return nil, err
		}
		asignaciones = append(asignaciones, asignacion)
	}
	return asignaciones, filas.Err()
}
