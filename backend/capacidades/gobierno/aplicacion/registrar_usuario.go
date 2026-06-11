package aplicacion

import (
	"context"

	"minas/capacidades/gobierno/dominio"
	"minas/capacidades/gobierno/puertos"
	"minas/compartido/identificador"
)

type ComandoRegistrarUsuario struct {
	IdentificadorEmpresa  string
	NombreCorto           string
	Nombre                string
	Correo                string
	IdentificadorEmpleado string
}

type RegistrarUsuario struct {
	unidad   puertos.UnidadDeTrabajo
	usuarios puertos.RepositorioUsuario
}

func NuevoRegistrarUsuario(unidad puertos.UnidadDeTrabajo, usuarios puertos.RepositorioUsuario) *RegistrarUsuario {
	return &RegistrarUsuario{unidad: unidad, usuarios: usuarios}
}

func (caso *RegistrarUsuario) Ejecutar(ctx context.Context, comando ComandoRegistrarUsuario) (string, error) {
	empresa, err := identificador.Desde(comando.IdentificadorEmpresa)
	if err != nil {
		return "", err
	}
	usuario, err := dominio.RegistrarUsuario(empresa, comando.NombreCorto, comando.Nombre)
	if err != nil {
		return "", err
	}
	if comando.Correo != "" {
		correo, err := dominio.CorreoDesde(comando.Correo)
		if err != nil {
			return "", err
		}
		usuario.DefinirCorreo(correo)
	}
	if comando.IdentificadorEmpleado != "" {
		empleado, err := identificador.Desde(comando.IdentificadorEmpleado)
		if err != nil {
			return "", err
		}
		usuario.VincularConEmpleado(empleado)
	}
	err = caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		existente, encontrado, err := caso.usuarios.BuscarPorNombreCorto(ctx, usuario.NombreCorto())
		if err != nil {
			return err
		}
		if encontrado && existente.EstaActivo() {
			return ErrUsuarioDuplicado
		}
		return caso.usuarios.Guardar(ctx, usuario)
	})
	if err != nil {
		return "", err
	}
	return usuario.Identificador().Texto(), nil
}
