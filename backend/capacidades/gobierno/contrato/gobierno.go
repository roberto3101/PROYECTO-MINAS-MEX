package contrato

import "context"

type PermisoVigente struct {
	Codigo      string
	Modulo      string
	AlcanceMina string
}

type ResumenEmpresa struct {
	Identificador string
	Codigo        string
	RazonSocial   string
	Activa        bool
	ZonaHoraria   string
	Moneda        string
	LogoUrl       string
	ColorPrimario string
	IdentificacionFiscal string
	CorreoContacto       string
	Telefono             string
}

type Gobierno interface {
	PermisosVigentesDe(ctx context.Context, identificadorUsuario string) ([]PermisoVigente, error)
	EmpresaActual(ctx context.Context) (ResumenEmpresa, bool, error)
}
