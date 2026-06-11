package puertos

import (
	"context"

	"minas/capacidades/gobierno/dominio"
	"minas/compartido/identificador"
)

type RepositorioEmpresa interface {
	BuscarPorIdentificador(ctx context.Context, id identificador.Identificador) (dominio.Empresa, bool, error)
	Guardar(ctx context.Context, empresa dominio.Empresa) error
}

type RepositorioUsuario interface {
	BuscarPorIdentificador(ctx context.Context, id identificador.Identificador) (dominio.Usuario, bool, error)
	BuscarPorNombreCorto(ctx context.Context, nombreCorto string) (dominio.Usuario, bool, error)
	Guardar(ctx context.Context, usuario dominio.Usuario) error
}

type RepositorioRol interface {
	BuscarPorIdentificador(ctx context.Context, id identificador.Identificador) (dominio.Rol, bool, error)
	BuscarPorCodigo(ctx context.Context, codigo string) (dominio.Rol, bool, error)
	Guardar(ctx context.Context, rol dominio.Rol) error
}

type RepositorioAsignacionRol interface {
	BuscarPorIdentificador(ctx context.Context, id identificador.Identificador) (dominio.AsignacionRol, bool, error)
	BuscarVigente(ctx context.Context, idUsuario, idRol identificador.Identificador, alcanceMina *identificador.Identificador) (dominio.AsignacionRol, bool, error)
	Guardar(ctx context.Context, asignacion dominio.AsignacionRol) error
}

type RepositorioPermiso interface {
	BuscarPorCodigo(ctx context.Context, codigo string) (dominio.Permiso, bool, error)
	Listar(ctx context.Context) ([]dominio.Permiso, error)
}
