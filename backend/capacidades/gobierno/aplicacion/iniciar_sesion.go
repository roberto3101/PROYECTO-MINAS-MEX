package aplicacion

import (
	"context"
	"errors"

	"minas/capacidades/gobierno/puertos"
)

var ErrCredencialesInvalidas = errors.New("credenciales invalidas")

type ComandoIniciarSesion struct {
	CodigoEmpresa string
	NombreCorto   string
	Contrasena    string
}

type SesionAutenticada struct {
	IdentificadorUsuario string
	IdentificadorEmpresa string
	NombreCorto          string
	Permisos             []string
}

type IniciarSesion struct {
	unidad   puertos.UnidadDeTrabajoDePlataforma
	lector   puertos.LectorDeAcceso
	cifrador puertos.CifradorDeContrasena
}

func NuevoIniciarSesion(unidad puertos.UnidadDeTrabajoDePlataforma, lector puertos.LectorDeAcceso, cifrador puertos.CifradorDeContrasena) *IniciarSesion {
	return &IniciarSesion{unidad: unidad, lector: lector, cifrador: cifrador}
}

func (caso *IniciarSesion) Ejecutar(ctx context.Context, comando ComandoIniciarSesion) (SesionAutenticada, error) {
	var sesion SesionAutenticada
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		credencial, encontrada, err := caso.lector.BuscarCredencial(ctx, comando.CodigoEmpresa, comando.NombreCorto)
		if err != nil {
			return err
		}
		if !encontrada || !credencial.UsuarioActivo || !caso.cifrador.Verificar(comando.Contrasena, credencial.HashContrasena) {
			return ErrCredencialesInvalidas
		}
		permisos, err := caso.lector.PermisosDe(ctx, credencial.IdentificadorUsuario)
		if err != nil {
			return err
		}
		sesion = SesionAutenticada{
			IdentificadorUsuario: credencial.IdentificadorUsuario,
			IdentificadorEmpresa: credencial.IdentificadorEmpresa,
			NombreCorto:          credencial.NombreCorto,
			Permisos:             permisos,
		}
		return nil
	})
	if err != nil {
		return SesionAutenticada{}, err
	}
	return sesion, nil
}
