package puertos

import "context"

type Credencial struct {
	IdentificadorUsuario string
	IdentificadorEmpresa string
	NombreCorto          string
	HashContrasena       string
	UsuarioActivo        bool
}

type LectorDeAcceso interface {
	BuscarCredencial(ctx context.Context, codigoEmpresa, nombreCorto string) (Credencial, bool, error)
	PermisosDe(ctx context.Context, identificadorUsuario string) ([]string, error)
}

type UnidadDeTrabajoDePlataforma interface {
	EnTransaccion(ctx context.Context, operacion func(ctx context.Context) error) error
}
