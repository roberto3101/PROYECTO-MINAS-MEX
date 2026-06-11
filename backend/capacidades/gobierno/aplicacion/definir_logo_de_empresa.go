package aplicacion

import (
	"context"

	"minas/capacidades/gobierno/puertos"
	"minas/compartido/identificador"
	"minas/plataforma/escudo"
)

type ComandoDefinirLogo struct {
	IdentificadorEmpresa string
	Datos                []byte
}

type DefinirLogoDeEmpresa struct {
	unidad   puertos.UnidadDeTrabajo
	empresas puertos.RepositorioEmpresa
	almacen  puertos.AlmacenDeArchivos
}

func NuevoDefinirLogoDeEmpresa(unidad puertos.UnidadDeTrabajo, empresas puertos.RepositorioEmpresa, almacen puertos.AlmacenDeArchivos) *DefinirLogoDeEmpresa {
	return &DefinirLogoDeEmpresa{unidad: unidad, empresas: empresas, almacen: almacen}
}

func (caso *DefinirLogoDeEmpresa) Ejecutar(ctx context.Context, comando ComandoDefinirLogo) (string, error) {
	idEmpresa, err := identificador.Desde(comando.IdentificadorEmpresa)
	if err != nil {
		return "", err
	}
	formato, err := escudo.ValidarImagenDeLogo(comando.Datos)
	if err != nil {
		return "", err
	}
	ruta, err := caso.almacen.GuardarLogo(idEmpresa.Texto(), formato, comando.Datos)
	if err != nil {
		return "", err
	}
	err = caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		empresa, encontrada, err := caso.empresas.BuscarPorIdentificador(ctx, idEmpresa)
		if err != nil {
			return err
		}
		if !encontrada {
			return ErrEmpresaNoEncontrada
		}
		empresa.DefinirLogo(ruta)
		return caso.empresas.Guardar(ctx, empresa)
	})
	if err != nil {
		return "", err
	}
	return ruta, nil
}
