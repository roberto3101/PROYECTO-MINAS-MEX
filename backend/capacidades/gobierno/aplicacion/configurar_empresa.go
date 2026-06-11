package aplicacion

import (
	"context"

	"minas/capacidades/gobierno/dominio"
	"minas/capacidades/gobierno/puertos"
	"minas/compartido/identificador"
)

type ComandoConfigurarEmpresa struct {
	IdentificadorEmpresa string
	LogoUrl              string
	ColorPrimario        string
	ZonaHoraria          string
	Moneda               string
}

type ConfigurarEmpresa struct {
	unidad   puertos.UnidadDeTrabajo
	empresas puertos.RepositorioEmpresa
}

func NuevoConfigurarEmpresa(unidad puertos.UnidadDeTrabajo, empresas puertos.RepositorioEmpresa) *ConfigurarEmpresa {
	return &ConfigurarEmpresa{unidad: unidad, empresas: empresas}
}

func (caso *ConfigurarEmpresa) Ejecutar(ctx context.Context, comando ComandoConfigurarEmpresa) error {
	identificadorEmpresa, err := identificador.Desde(comando.IdentificadorEmpresa)
	if err != nil {
		return err
	}
	branding, err := dominio.ConfigurarBranding(comando.LogoUrl, comando.ColorPrimario, comando.ZonaHoraria, comando.Moneda)
	if err != nil {
		return err
	}
	return caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		empresa, encontrada, err := caso.empresas.BuscarPorIdentificador(ctx, identificadorEmpresa)
		if err != nil {
			return err
		}
		if !encontrada {
			return ErrEmpresaNoEncontrada
		}
		empresa.AplicarBranding(branding)
		return caso.empresas.Guardar(ctx, empresa)
	})
}
