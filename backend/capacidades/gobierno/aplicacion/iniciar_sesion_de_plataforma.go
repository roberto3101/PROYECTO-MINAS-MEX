package aplicacion

import (
	"context"
	"errors"

	"minas/capacidades/gobierno/puertos"
	"minas/compartido/reloj"
	"minas/plataforma/escudo"
)

type ComandoIniciarSesionDePlataforma struct {
	NombreCorto       string
	Contrasena        string
	ClaveDeLimitacion string
}

type SesionDePlataforma struct {
	IdentificadorSuperadmin string
	NombreCorto             string
}

type IniciarSesionDePlataforma struct {
	unidad    puertos.UnidadDeTrabajoDePlataforma
	lector    puertos.LectorDePlataforma
	cifrador  puertos.CifradorDeContrasena
	limitador *escudo.LimitadorDeIntentos
	reloj     reloj.Reloj
}

func NuevoIniciarSesionDePlataforma(unidad puertos.UnidadDeTrabajoDePlataforma, lector puertos.LectorDePlataforma, cifrador puertos.CifradorDeContrasena, limitador *escudo.LimitadorDeIntentos, relojDelSistema reloj.Reloj) *IniciarSesionDePlataforma {
	return &IniciarSesionDePlataforma{unidad: unidad, lector: lector, cifrador: cifrador, limitador: limitador, reloj: relojDelSistema}
}

func (caso *IniciarSesionDePlataforma) Ejecutar(ctx context.Context, comando ComandoIniciarSesionDePlataforma) (SesionDePlataforma, error) {
	ahora := caso.reloj.Ahora()
	if espera, permitido := caso.limitador.Permitir(comando.ClaveDeLimitacion, ahora); !permitido {
		return SesionDePlataforma{}, &ErrDemasiadosIntentos{EsperaSegundos: espera}
	}
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
		if errors.Is(err, ErrCredencialesInvalidas) {
			caso.limitador.RegistrarFallo(comando.ClaveDeLimitacion, ahora)
		}
		return SesionDePlataforma{}, err
	}
	caso.limitador.RegistrarExito(comando.ClaveDeLimitacion)
	return sesion, nil
}
