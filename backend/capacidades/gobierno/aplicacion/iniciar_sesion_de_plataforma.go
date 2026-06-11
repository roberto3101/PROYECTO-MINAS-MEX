package aplicacion

import (
	"context"

	"minas/capacidades/gobierno/puertos"
)

type ComandoIniciarSesionDePlataforma struct {
	NombreCorto string
	Contrasena  string
}

type SesionDePlataforma struct {
	IdentificadorSuperadmin string
	NombreCorto             string
}

type IniciarSesionDePlataforma struct {
	unidad   puertos.UnidadDeTrabajoDePlataforma
	lector   puertos.LectorDePlataforma
	cifrador puertos.CifradorDeContrasena
}

func NuevoIniciarSesionDePlataforma(unidad puertos.UnidadDeTrabajoDePlataforma, lector puertos.LectorDePlataforma, cifrador puertos.CifradorDeContrasena) *IniciarSesionDePlataforma {
	return &IniciarSesionDePlataforma{unidad: unidad, lector: lector, cifrador: cifrador}
}

func (caso *IniciarSesionDePlataforma) Ejecutar(ctx context.Context, comando ComandoIniciarSesionDePlataforma) (SesionDePlataforma, error) {
	var sesion SesionDePlataforma
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		credencial, encontrada, err := caso.lector.BuscarCredencialDeSuperadmin(ctx, comando.NombreCorto)
		if err != nil {
			return err
		}
		if !encontrada || !credencial.Activo || credencial.HashContrasena == "" ||
			!caso.cifrador.Verificar(comando.Contrasena, credencial.HashContrasena) {
			return ErrCredencialesInvalidas
		}
		sesion = SesionDePlataforma{IdentificadorSuperadmin: credencial.Identificador, NombreCorto: credencial.NombreCorto}
		return nil
	})
	if err != nil {
		return SesionDePlataforma{}, err
	}
	return sesion, nil
}
