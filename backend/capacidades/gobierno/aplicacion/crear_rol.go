package aplicacion

import (
	"context"

	"minas/capacidades/gobierno/dominio"
	"minas/capacidades/gobierno/puertos"
	"minas/compartido/identificador"
)

type ComandoCrearRol struct {
	IdentificadorEmpresa string
	Codigo               string
	Descripcion          string
}

type CrearRol struct {
	unidad puertos.UnidadDeTrabajo
	roles  puertos.RepositorioRol
}

func NuevoCrearRol(unidad puertos.UnidadDeTrabajo, roles puertos.RepositorioRol) *CrearRol {
	return &CrearRol{unidad: unidad, roles: roles}
}

func (caso *CrearRol) Ejecutar(ctx context.Context, comando ComandoCrearRol) (string, error) {
	empresa, err := identificador.Desde(comando.IdentificadorEmpresa)
	if err != nil {
		return "", err
	}
	rol, err := dominio.CrearRolPropio(empresa, comando.Codigo, comando.Descripcion)
	if err != nil {
		return "", err
	}
	err = caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		return caso.roles.Guardar(ctx, rol)
	})
	if err != nil {
		return "", err
	}
	return rol.Identificador().Texto(), nil
}
