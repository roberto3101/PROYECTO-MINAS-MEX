package infraestructura

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"minas/capacidades/gobierno/dominio"
	"minas/compartido/identificador"
	"minas/plataforma/persistencia"
)

type RepositorioUsuarioPostgres struct{}

func NuevoRepositorioUsuario() RepositorioUsuarioPostgres {
	return RepositorioUsuarioPostgres{}
}

func (RepositorioUsuarioPostgres) Guardar(ctx context.Context, usuario dominio.Usuario) error {
	consultas := persistencia.ConsultasDe(ctx)
	_, err := consultas.Exec(ctx,
		`INSERT INTO gobierno.usuario (id, id_empresa, usuario, nombre, correo, id_empleado, estado)
		 VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, $7)
		 ON CONFLICT (id) DO UPDATE SET
		   nombre = EXCLUDED.nombre,
		   correo = EXCLUDED.correo,
		   id_empleado = EXCLUDED.id_empleado,
		   estado = EXCLUDED.estado,
		   actualizado_en = now()`,
		usuario.Identificador().Texto(), usuario.Empresa().Texto(), usuario.NombreCorto(),
		usuario.Nombre(), usuario.Correo(), textoOpcional(usuario.EmpleadoVinculado()), string(usuario.Estado()))
	return err
}

func (RepositorioUsuarioPostgres) BuscarPorIdentificador(ctx context.Context, id identificador.Identificador) (dominio.Usuario, bool, error) {
	return leerUsuario(ctx, "WHERE id = $1 AND eliminado_en IS NULL", id.Texto())
}

func (RepositorioUsuarioPostgres) BuscarPorNombreCorto(ctx context.Context, nombreCorto string) (dominio.Usuario, bool, error) {
	return leerUsuario(ctx, "WHERE usuario = $1 AND eliminado_en IS NULL", nombreCorto)
}

func leerUsuario(ctx context.Context, filtro, argumento string) (dominio.Usuario, bool, error) {
	consultas := persistencia.ConsultasDe(ctx)
	var id, idEmpresa, nombreCorto, nombre, correo, estado string
	var idEmpleado *string
	fila := consultas.QueryRow(ctx,
		"SELECT id, id_empresa, usuario, nombre, COALESCE(correo, ''), id_empleado, estado FROM gobierno.usuario "+filtro,
		argumento)
	if err := fila.Scan(&id, &idEmpresa, &nombreCorto, &nombre, &correo, &idEmpleado, &estado); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dominio.Usuario{}, false, nil
		}
		return dominio.Usuario{}, false, err
	}
	identificadorUsuario, err := identificador.Desde(id)
	if err != nil {
		return dominio.Usuario{}, false, err
	}
	identificadorEmpresa, err := identificador.Desde(idEmpresa)
	if err != nil {
		return dominio.Usuario{}, false, err
	}
	empleado, err := identificadorOpcional(idEmpleado)
	if err != nil {
		return dominio.Usuario{}, false, err
	}
	return dominio.ReconstruirUsuario(identificadorUsuario, identificadorEmpresa, nombreCorto, nombre, correo, empleado, dominio.EstadoUsuario(estado)), true, nil
}
