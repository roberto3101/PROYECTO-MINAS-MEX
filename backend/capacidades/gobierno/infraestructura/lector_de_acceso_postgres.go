package infraestructura

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"minas/capacidades/gobierno/puertos"
	"minas/plataforma/persistencia"
)

type LectorDeAccesoPostgres struct{}

func NuevoLectorDeAcceso() LectorDeAccesoPostgres {
	return LectorDeAccesoPostgres{}
}

func (LectorDeAccesoPostgres) BuscarCredencial(ctx context.Context, codigoEmpresa, nombreCorto string) (puertos.Credencial, bool, error) {
	consultas := persistencia.ConsultasDe(ctx)
	var credencial puertos.Credencial
	fila := consultas.QueryRow(ctx,
		`SELECT u.id, u.id_empresa, u.usuario, COALESCE(u.contrasena_hash, ''), (u.estado = 'ACTIVO' AND u.eliminado_en IS NULL)
		 FROM gobierno.usuario u
		 JOIN gobierno.empresa e ON e.id = u.id_empresa
		 WHERE e.codigo = $1 AND u.usuario = $2 AND e.eliminado_en IS NULL AND e.estado = 'ACTIVA'`,
		codigoEmpresa, nombreCorto)
	if err := fila.Scan(&credencial.IdentificadorUsuario, &credencial.IdentificadorEmpresa, &credencial.NombreCorto, &credencial.HashContrasena, &credencial.UsuarioActivo); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return puertos.Credencial{}, false, nil
		}
		return puertos.Credencial{}, false, err
	}
	return credencial, true, nil
}

func (LectorDeAccesoPostgres) PermisosDe(ctx context.Context, identificadorUsuario string) ([]string, error) {
	consultas := persistencia.ConsultasDe(ctx)
	filas, err := consultas.Query(ctx, "SELECT permiso FROM gobierno.v_permisos_usuario WHERE id_usuario = $1", identificadorUsuario)
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	var permisos []string
	for filas.Next() {
		var permiso string
		if err := filas.Scan(&permiso); err != nil {
			return nil, err
		}
		permisos = append(permisos, permiso)
	}
	return permisos, filas.Err()
}

func (LectorDeAccesoPostgres) AlcanceDeMinas(ctx context.Context, identificadorUsuario string) (bool, []string, error) {
	consultas := persistencia.ConsultasDe(ctx)
	var esGlobal bool
	err := consultas.QueryRow(ctx,
		`SELECT EXISTS (
		   SELECT 1 FROM gobierno.usuario_rol ur
		   JOIN gobierno.rol r ON r.id = ur.id_rol AND r.eliminado_en IS NULL AND r.estado = 'ACTIVO'
		   WHERE ur.id_usuario = $1 AND ur.eliminado_en IS NULL AND ur.id_mina IS NULL)`,
		identificadorUsuario).Scan(&esGlobal)
	if err != nil {
		return false, nil, err
	}
	if esGlobal {
		return true, nil, nil
	}
	filas, err := consultas.Query(ctx,
		`SELECT DISTINCT ur.id_mina::text FROM gobierno.usuario_rol ur
		 JOIN gobierno.rol r ON r.id = ur.id_rol AND r.eliminado_en IS NULL AND r.estado = 'ACTIVO'
		 WHERE ur.id_usuario = $1 AND ur.eliminado_en IS NULL AND ur.id_mina IS NOT NULL`,
		identificadorUsuario)
	if err != nil {
		return false, nil, err
	}
	defer filas.Close()
	var minas []string
	for filas.Next() {
		var mina string
		if err := filas.Scan(&mina); err != nil {
			return false, nil, err
		}
		minas = append(minas, mina)
	}
	return false, minas, filas.Err()
}
