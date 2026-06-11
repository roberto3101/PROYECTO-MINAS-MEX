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

func (RepositorioEmpresaPostgres) BuscarPorIdentificador(ctx context.Context, id identificador.Identificador) (dominio.Empresa, bool, error) {
	consultas := persistencia.ConsultasDe(ctx)
	fila := consultas.QueryRow(ctx,
		`SELECT id, codigo, razon_social,
		        COALESCE(identificacion_fiscal,''), COALESCE(correo_contacto,''), COALESCE(telefono,''),
		        COALESCE(logo_url,''), COALESCE(color_primario,''), zona_horaria, moneda_defecto, estado
		 FROM gobierno.empresa WHERE id = $1 AND eliminado_en IS NULL`, id.Texto())
	var crudoId, codigo, razon, fiscal, correo, telefono, logo, color, zona, moneda, estado string
	if err := fila.Scan(&crudoId, &codigo, &razon, &fiscal, &correo, &telefono, &logo, &color, &zona, &moneda, &estado); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dominio.Empresa{}, false, nil
		}
		return dominio.Empresa{}, false, err
	}
	idEmpresa, err := identificador.Desde(crudoId)
	if err != nil {
		return dominio.Empresa{}, false, err
	}
	branding, err := dominio.ConfigurarBranding(logo, color, zona, moneda)
	if err != nil {
		return dominio.Empresa{}, false, err
	}
	perfil := dominio.PerfilDeContacto{IdentificacionFiscal: fiscal, CorreoContacto: correo, Telefono: telefono}
	return dominio.ReconstruirEmpresa(idEmpresa, codigo, razon, perfil, branding, dominio.EstadoEmpresa(estado)), true, nil
}

func (RepositorioEmpresaPostgres) Guardar(ctx context.Context, empresa dominio.Empresa) error {
	consultas := persistencia.ConsultasDe(ctx)
	branding := empresa.Branding()
	perfil := empresa.Perfil()
	_, err := consultas.Exec(ctx,
		`UPDATE gobierno.empresa SET
		   razon_social = $2, identificacion_fiscal = NULLIF($3,''), correo_contacto = NULLIF($4,''),
		   telefono = NULLIF($5,''), logo_url = NULLIF($6,''), color_primario = NULLIF($7,''),
		   zona_horaria = $8, moneda_defecto = $9, actualizado_en = now()
		 WHERE id = $1`,
		empresa.Identificador().Texto(), empresa.RazonSocial(),
		perfil.IdentificacionFiscal, perfil.CorreoContacto, perfil.Telefono,
		branding.LogoUrl(), branding.ColorPrimario(), branding.ZonaHoraria(), branding.Moneda())
	return err
}

func (RepositorioEmpresaPostgres) CambiarEstado(ctx context.Context, id identificador.Identificador, estado string) error {
	consultas := persistencia.ConsultasDe(ctx)
	_, err := consultas.Exec(ctx,
		`UPDATE gobierno.empresa SET estado = $2, actualizado_en = now() WHERE id = $1`,
		id.Texto(), estado)
	return err
}

func (RepositorioEmpresaPostgres) Crear(ctx context.Context, empresa dominio.Empresa) error {
	consultas := persistencia.ConsultasDe(ctx)
	branding := empresa.Branding()
	perfil := empresa.Perfil()
	_, err := consultas.Exec(ctx,
		`INSERT INTO gobierno.empresa
		   (id, codigo, razon_social, identificacion_fiscal, correo_contacto, telefono,
		    logo_url, color_primario, zona_horaria, moneda_defecto, estado)
		 VALUES ($1,$2,$3,NULLIF($4,''),NULLIF($5,''),NULLIF($6,''),NULLIF($7,''),NULLIF($8,''),$9,$10,$11)`,
		empresa.Identificador().Texto(), empresa.Codigo(), empresa.RazonSocial(),
		perfil.IdentificacionFiscal, perfil.CorreoContacto, perfil.Telefono,
		branding.LogoUrl(), branding.ColorPrimario(), branding.ZonaHoraria(), branding.Moneda(),
		string(empresa.Estado()))
	return err
}
