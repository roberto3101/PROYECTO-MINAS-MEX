package infraestructura

import (
	"context"

	"minas/capacidades/gobierno/puertos"
	"minas/compartido/paginacion"
	"minas/plataforma/persistencia"
)

type LectorDeGobiernoPostgres struct{}

func NuevoLectorDeGobierno() LectorDeGobiernoPostgres {
	return LectorDeGobiernoPostgres{}
}

func (LectorDeGobiernoPostgres) ListarUsuarios(ctx context.Context, filtro puertos.FiltroDeUsuarios) ([]puertos.ResumenUsuario, string, error) {
	consultas := persistencia.ConsultasDe(ctx)
	orden, identificadorCursor, err := paginacion.DecodificarCursor(filtro.Cursor)
	if err != nil {
		return nil, "", err
	}
	filas, err := consultas.Query(ctx,
		`SELECT id, usuario, nombre, COALESCE(correo, ''), estado,
		        COALESCE(to_char(ultimo_acceso_en, 'YYYY-MM-DD"T"HH24:MI:SSZ'), '')
		 FROM gobierno.usuario
		 WHERE eliminado_en IS NULL
		   AND ($1 = '' OR usuario ILIKE '%'||$1||'%' OR nombre ILIKE '%'||$1||'%' OR COALESCE(correo,'') ILIKE '%'||$1||'%')
		   AND ($2 = '' OR $2 = 'TODOS' OR estado = $2)
		   AND ($3 = '' OR (usuario, id::text) > ($3, $4))
		 ORDER BY usuario, id LIMIT $5`,
		filtro.Busqueda, filtro.Estado, orden, identificadorCursor, filtro.Limite+1)
	if err != nil {
		return nil, "", err
	}
	defer filas.Close()
	var usuarios []puertos.ResumenUsuario
	for filas.Next() {
		var usuario puertos.ResumenUsuario
		if err := filas.Scan(&usuario.Identificador, &usuario.NombreCorto, &usuario.Nombre, &usuario.Correo, &usuario.Estado, &usuario.UltimoAccesoEn); err != nil {
			return nil, "", err
		}
		usuarios = append(usuarios, usuario)
	}
	if err := filas.Err(); err != nil {
		return nil, "", err
	}
	siguiente := ""
	if len(usuarios) > filtro.Limite {
		usuarios = usuarios[:filtro.Limite]
		ultimo := usuarios[len(usuarios)-1]
		siguiente = paginacion.CodificarCursor(ultimo.NombreCorto, ultimo.Identificador)
	}
	return usuarios, siguiente, nil
}

func (LectorDeGobiernoPostgres) DetalleDeUsuario(ctx context.Context, identificadorUsuario string) (puertos.DetalleUsuario, bool, error) {
	consultas := persistencia.ConsultasDe(ctx)
	var detalle puertos.DetalleUsuario
	fila := consultas.QueryRow(ctx,
		`SELECT id, usuario, nombre, COALESCE(correo, ''), estado,
		        COALESCE(to_char(ultimo_acceso_en, 'YYYY-MM-DD"T"HH24:MI:SSZ'), ''),
		        to_char(creado_en, 'YYYY-MM-DD"T"HH24:MI:SSZ'), COALESCE(id_empleado::text, '')
		 FROM gobierno.usuario WHERE id = $1 AND eliminado_en IS NULL`, identificadorUsuario)
	err := fila.Scan(&detalle.Identificador, &detalle.NombreCorto, &detalle.Nombre, &detalle.Correo,
		&detalle.Estado, &detalle.UltimoAccesoEn, &detalle.CreadoEn, &detalle.IdentificadorEmpleado)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return puertos.DetalleUsuario{}, false, nil
		}
		return puertos.DetalleUsuario{}, false, err
	}
	return detalle, true, nil
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
