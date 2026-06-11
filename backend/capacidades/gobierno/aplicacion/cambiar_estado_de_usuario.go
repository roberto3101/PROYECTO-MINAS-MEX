package aplicacion

import (
	"context"

	"minas/capacidades/gobierno/dominio"
	"minas/capacidades/gobierno/puertos"
	"minas/compartido/identificador"
)

type ComandoCambiarEstadoDeUsuario struct {
	IdentificadorUsuario string
	Estado               string
}

type CambiarEstadoDeUsuario struct {
	unidad   puertos.UnidadDeTrabajo
	usuarios puertos.RepositorioUsuario
}

func NuevoCambiarEstadoDeUsuario(unidad puertos.UnidadDeTrabajo, usuarios puertos.RepositorioUsuario) *CambiarEstadoDeUsuario {
	return &CambiarEstadoDeUsuario{unidad: unidad, usuarios: usuarios}
}

func (caso *CambiarEstadoDeUsuario) Ejecutar(ctx context.Context, comando ComandoCambiarEstadoDeUsuario) error {
	idUsuario, err := identificador.Desde(comando.IdentificadorUsuario)
	if err != nil {
		return err
	}
	if comando.Estado != string(dominio.UsuarioActivo) && comando.Estado != string(dominio.UsuarioInactivo) {
		return dominio.ErrEstadoNoReconocido
	}
	return caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		usuario, encontrado, err := caso.usuarios.BuscarPorIdentificador(ctx, idUsuario)
		if err != nil {
			return err
		}
		if !encontrado {
			return ErrUsuarioNoEncontrado
		}
		if comando.Estado == string(dominio.UsuarioActivo) {
			err = usuario.Reactivar()
		} else {
			err = usuario.Desactivar()
		}
		if err != nil {
			return err
		}
		return caso.usuarios.Guardar(ctx, usuario)
	})
}
