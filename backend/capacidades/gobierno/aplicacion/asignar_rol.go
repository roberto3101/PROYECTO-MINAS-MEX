package aplicacion

import (
	"context"

	"minas/capacidades/gobierno/dominio"
	"minas/capacidades/gobierno/puertos"
	"minas/compartido/identificador"
)

type ComandoAsignarRol struct {
	IdentificadorEmpresa string
	IdentificadorUsuario string
	IdentificadorRol     string
	AlcanceMina          string
}

type AsignarRol struct {
	unidad       puertos.UnidadDeTrabajo
	asignaciones puertos.RepositorioAsignacionRol
}

func NuevoAsignarRol(unidad puertos.UnidadDeTrabajo, asignaciones puertos.RepositorioAsignacionRol) *AsignarRol {
	return &AsignarRol{unidad: unidad, asignaciones: asignaciones}
}

func (caso *AsignarRol) Ejecutar(ctx context.Context, comando ComandoAsignarRol) (string, error) {
	empresa, err := identificador.Desde(comando.IdentificadorEmpresa)
	if err != nil {
		return "", err
	}
	usuario, err := identificador.Desde(comando.IdentificadorUsuario)
	if err != nil {
		return "", err
	}
	rol, err := identificador.Desde(comando.IdentificadorRol)
	if err != nil {
		return "", err
	}
	alcance, err := identificador.DesdeOpcional(comando.AlcanceMina)
	if err != nil {
		return "", err
	}
	asignacion := dominio.AsignarRol(empresa, usuario, rol, alcance)
	err = caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		return caso.asignaciones.Guardar(ctx, asignacion)
	})
	if err != nil {
		return "", err
	}
	return asignacion.Identificador().Texto(), nil
}
