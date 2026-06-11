package aplicacion

import (
	"context"

	"minas/capacidades/gobierno/dominio"
	"minas/capacidades/gobierno/puertos"
	"minas/compartido/identificador"
	"minas/plataforma/escudo"
)

type ComandoEditarUsuario struct {
	IdentificadorUsuario string
	Nombre               string
	Correo               string
}

type EditarUsuario struct {
	unidad   puertos.UnidadDeTrabajo
	usuarios puertos.RepositorioUsuario
}

func NuevoEditarUsuario(unidad puertos.UnidadDeTrabajo, usuarios puertos.RepositorioUsuario) *EditarUsuario {
	return &EditarUsuario{unidad: unidad, usuarios: usuarios}
}

func (caso *EditarUsuario) Ejecutar(ctx context.Context, comando ComandoEditarUsuario) error {
	idUsuario, err := identificador.Desde(comando.IdentificadorUsuario)
	if err != nil {
		return err
	}
	if err := escudo.ValidarCorreo(comando.Correo); err != nil {
		return err
	}
	return caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		usuario, encontrado, err := caso.usuarios.BuscarPorIdentificador(ctx, idUsuario)
		if err != nil {
			return err
		}
		if !encontrado {
			return ErrUsuarioNoEncontrado
		}
		if err := usuario.Renombrar(comando.Nombre); err != nil {
			return err
		}
		if comando.Correo != "" {
			correo, err := dominio.NuevoCorreo(comando.Correo)
			if err != nil {
				return err
			}
			usuario.DefinirCorreo(correo)
		}
		return caso.usuarios.Guardar(ctx, usuario)
	})
}
