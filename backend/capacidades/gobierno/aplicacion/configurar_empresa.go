package aplicacion

import (
	"context"

	"minas/capacidades/gobierno/dominio"
	"minas/capacidades/gobierno/puertos"
	"minas/compartido/identificador"
	"minas/plataforma/escudo"
)

type ComandoConfigurarEmpresa struct {
	IdentificadorEmpresa string
	ColorPrimario        string
	ZonaHoraria          string
	Moneda               string
	IdentificacionFiscal string
	CorreoContacto       string
	Telefono             string
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
	if err := escudo.ValidarCorreo(comando.CorreoContacto); err != nil {
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
		branding, err := dominio.ConfigurarBranding(empresa.Branding().LogoUrl(), comando.ColorPrimario, comando.ZonaHoraria, comando.Moneda)
		if err != nil {
			return err
		}
		empresa.AplicarBranding(branding)
		empresa.ActualizarPerfil(dominio.PerfilDeContacto{
			IdentificacionFiscal: comando.IdentificacionFiscal,
			CorreoContacto:       comando.CorreoContacto,
			Telefono:             comando.Telefono,
		})
		return caso.empresas.Guardar(ctx, empresa)
	})
}
