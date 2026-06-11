package infraestructura

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"minas/capacidades/gobierno/dominio"
	"minas/compartido/identificador"
	"minas/plataforma/persistencia"
)

type RepositorioEmpresaPostgres struct{}

func NuevoRepositorioEmpresa() RepositorioEmpresaPostgres {
	return RepositorioEmpresaPostgres{}
}

func (RepositorioEmpresaPostgres) Guardar(ctx context.Context, empresa dominio.Empresa) error {
	consultas := persistencia.ConsultasDe(ctx)
	branding := empresa.Branding()
	_, err := consultas.Exec(ctx,
		`UPDATE gobierno.empresa SET
		   logo_url = NULLIF($2, ''),
		   color_primario = NULLIF($3, ''),
		   zona_horaria = $4,
		   moneda_defecto = $5,
		   actualizado_en = now()
		 WHERE id = $1`,
		empresa.Identificador().Texto(), branding.LogoUrl(), branding.ColorPrimario(), branding.ZonaHoraria(), branding.Moneda())
	return err
}

func (RepositorioEmpresaPostgres) BuscarPorIdentificador(ctx context.Context, id identificador.Identificador) (dominio.Empresa, bool, error) {
	consultas := persistencia.ConsultasDe(ctx)
	var identificadorEmpresa, codigo, razonSocial, logoUrl, colorPrimario, zonaHoraria, moneda, estado string
	fila := consultas.QueryRow(ctx,
		`SELECT id, codigo, razon_social, COALESCE(logo_url, ''), COALESCE(color_primario, ''), zona_horaria, moneda_defecto, estado
		 FROM gobierno.empresa WHERE id = $1`, id.Texto())
	if err := fila.Scan(&identificadorEmpresa, &codigo, &razonSocial, &logoUrl, &colorPrimario, &zonaHoraria, &moneda, &estado); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dominio.Empresa{}, false, nil
		}
		return dominio.Empresa{}, false, err
	}
	convertido, err := identificador.Desde(identificadorEmpresa)
	if err != nil {
		return dominio.Empresa{}, false, err
	}
	branding, err := dominio.ConfigurarBranding(logoUrl, colorPrimario, zonaHoraria, moneda)
	if err != nil {
		return dominio.Empresa{}, false, err
	}
	return dominio.ReconstruirEmpresa(convertido, codigo, razonSocial, branding, dominio.EstadoEmpresa(estado)), true, nil
}
