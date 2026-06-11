package infraestructura

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"minas/capacidades/gobierno/dominio"
	"minas/compartido/identificador"
	"minas/plataforma/persistencia"
)

type RepositorioAsignacionPostgres struct{}

func NuevoRepositorioAsignacion() RepositorioAsignacionPostgres {
	return RepositorioAsignacionPostgres{}
}

func (RepositorioAsignacionPostgres) Guardar(ctx context.Context, asignacion dominio.AsignacionRol) error {
	consultas := persistencia.ConsultasDe(ctx)
	_, err := consultas.Exec(ctx,
		`INSERT INTO gobierno.usuario_rol (id, id_empresa, id_usuario, id_rol, id_mina)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (id) DO UPDATE SET
		   eliminado_en = CASE WHEN $6::boolean THEN now() ELSE NULL END,
		   actualizado_en = now()`,
		asignacion.Identificador().Texto(), asignacion.Empresa().Texto(), asignacion.Usuario().Texto(),
		asignacion.Rol().Texto(), textoOpcional(asignacion.AlcanceMina()), !asignacion.EstaVigente())
	return err
}

func (RepositorioAsignacionPostgres) BuscarPorIdentificador(ctx context.Context, id identificador.Identificador) (dominio.AsignacionRol, bool, error) {
	consultas := persistencia.ConsultasDe(ctx)
	return leerAsignacion(consultas.QueryRow(ctx,
		`SELECT id, id_empresa, id_usuario, id_rol, id_mina, (eliminado_en IS NOT NULL)
		 FROM gobierno.usuario_rol WHERE id = $1`, id.Texto()))
}

func (RepositorioAsignacionPostgres) BuscarVigente(ctx context.Context, idUsuario, idRol identificador.Identificador, alcanceMina *identificador.Identificador) (dominio.AsignacionRol, bool, error) {
	consultas := persistencia.ConsultasDe(ctx)
	return leerAsignacion(consultas.QueryRow(ctx,
		`SELECT id, id_empresa, id_usuario, id_rol, id_mina, (eliminado_en IS NOT NULL)
		 FROM gobierno.usuario_rol
		 WHERE id_usuario = $1 AND id_rol = $2 AND id_mina IS NOT DISTINCT FROM $3 AND eliminado_en IS NULL`,
		idUsuario.Texto(), idRol.Texto(), textoOpcional(alcanceMina)))
}

func leerAsignacion(fila pgx.Row) (dominio.AsignacionRol, bool, error) {
	var id, idEmpresa, idUsuario, idRol string
	var idMina *string
	var revocada bool
	if err := fila.Scan(&id, &idEmpresa, &idUsuario, &idRol, &idMina, &revocada); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dominio.AsignacionRol{}, false, nil
		}
		return dominio.AsignacionRol{}, false, err
	}
	identificadorAsignacion, err := identificador.Desde(id)
	if err != nil {
		return dominio.AsignacionRol{}, false, err
	}
	identificadorEmpresa, err := identificador.Desde(idEmpresa)
	if err != nil {
		return dominio.AsignacionRol{}, false, err
	}
	identificadorUsuario, err := identificador.Desde(idUsuario)
	if err != nil {
		return dominio.AsignacionRol{}, false, err
	}
	identificadorRol, err := identificador.Desde(idRol)
	if err != nil {
		return dominio.AsignacionRol{}, false, err
	}
	alcance, err := identificadorOpcional(idMina)
	if err != nil {
		return dominio.AsignacionRol{}, false, err
	}
	return dominio.ReconstruirAsignacion(identificadorAsignacion, identificadorEmpresa, identificadorUsuario, identificadorRol, alcance, revocada), true, nil
}
