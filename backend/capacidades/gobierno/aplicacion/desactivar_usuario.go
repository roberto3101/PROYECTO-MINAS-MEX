package aplicacion

import (
	"context"

	"minas/capacidades/gobierno/puertos"
	"minas/compartido/identificador"
)

type ComandoDesactivarUsuario struct {
	IdentificadorUsuario string
}

type DesactivarUsuario struct {
	unidad   puertos.UnidadDeTrabajo
	usuarios puertos.RepositorioUsuario
}

func NuevoDesactivarUsuario(unidad puertos.UnidadDeTrabajo, usuarios puertos.RepositorioUsuario) *DesactivarUsuario {
	return &DesactivarUsuario{unidad: unidad, usuarios: usuarios}
}

func (caso *DesactivarUsuario) Ejecutar(ctx context.Context, comando ComandoDesactivarUsuario) error {
	identificadorUsuario, err := identificador.Desde(comando.IdentificadorUsuario)
	if err != nil {
		return err
	}
	return caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		usuario, encontrado, err := caso.usuarios.BuscarPorIdentificador(ctx, identificadorUsuario)
		if err != nil {
			return err
		}
		if !encontrado {
			return ErrUsuarioNoEncontrado
		}
		if err := usuario.Desactivar(); err != nil {
			return err
		}
		return caso.usuarios.Guardar(ctx, usuario)
	})
}
