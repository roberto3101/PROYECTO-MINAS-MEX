package aplicacion

import (
	"context"
	"errors"

	"minas/capacidades/gobierno/puertos"
	"minas/compartido/reloj"
	"minas/plataforma/escudo"
)

var ErrCredencialesInvalidas = errors.New("usuario o contrasena incorrectos")

type ComandoIniciarSesion struct {
	CodigoEmpresa     string
	NombreCorto       string
	Contrasena        string
	ClaveDeLimitacion string
}

type SesionAutenticada struct {
	IdentificadorUsuario string
	IdentificadorEmpresa string
	NombreCorto          string
	Permisos             []string
	AlcanceGlobalDeMinas bool
	MinasPermitidas      []string
}

type IniciarSesion struct {
	unidad    puertos.UnidadDeTrabajoDePlataforma
	lector    puertos.LectorDeAcceso
	cifrador  puertos.CifradorDeContrasena
	limitador *escudo.LimitadorDeIntentos
	reloj     reloj.Reloj
}

func NuevoIniciarSesion(unidad puertos.UnidadDeTrabajoDePlataforma, lector puertos.LectorDeAcceso, cifrador puertos.CifradorDeContrasena, limitador *escudo.LimitadorDeIntentos, relojDelSistema reloj.Reloj) *IniciarSesion {
	return &IniciarSesion{unidad: unidad, lector: lector, cifrador: cifrador, limitador: limitador, reloj: relojDelSistema}
}

func (caso *IniciarSesion) Ejecutar(ctx context.Context, comando ComandoIniciarSesion) (SesionAutenticada, error) {
	ahora := caso.reloj.Ahora()
	if espera, permitido := caso.limitador.Permitir(comando.ClaveDeLimitacion, ahora); !permitido {
		return SesionAutenticada{}, &ErrDemasiadosIntentos{EsperaSegundos: espera}
	}
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
		esGlobal, minas, err := caso.lector.AlcanceDeMinas(ctx, credencial.IdentificadorUsuario)
		if err != nil {
			return err
		}
		sesion = SesionAutenticada{
			IdentificadorUsuario: credencial.IdentificadorUsuario,
			IdentificadorEmpresa: credencial.IdentificadorEmpresa,
			NombreCorto:          credencial.NombreCorto,
			Permisos:             permisos,
			AlcanceGlobalDeMinas: esGlobal,
			MinasPermitidas:      minas,
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrCredencialesInvalidas) {
			caso.limitador.RegistrarFallo(comando.ClaveDeLimitacion, ahora)
		}
		return SesionAutenticada{}, err
	}
	caso.limitador.RegistrarExito(comando.ClaveDeLimitacion)
	return sesion, nil
}
