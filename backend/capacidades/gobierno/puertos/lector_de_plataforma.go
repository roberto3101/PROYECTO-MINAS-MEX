package puertos

import "context"

type CredencialDeSuperadmin struct {
	Identificador  string
	NombreCorto    string
	Nombre         string
	HashContrasena string
	Activo         bool
}

type FiltroDeEmpresas struct {
	Busqueda string
	Estado   string
	Cursor   string
	Limite   int
}

type ResumenEmpresaDePlataforma struct {
	Identificador string
	Codigo        string
	RazonSocial   string
	Estado        string
	LogoUrl       string
	ColorPrimario string
	Moneda        string
	ZonaHoraria   string
	CreadoEn      string
	TotalUsuarios int
}

type DetalleEmpresaDePlataforma struct {
	ResumenEmpresaDePlataforma
	IdentificacionFiscal string
	CorreoContacto       string
	Telefono             string
	TotalMinas           int
}

type LectorDePlataforma interface {
	BuscarCredencialDeSuperadmin(ctx context.Context, nombreCorto string) (CredencialDeSuperadmin, bool, error)
	ListarEmpresas(ctx context.Context, filtro FiltroDeEmpresas) ([]ResumenEmpresaDePlataforma, string, error)
	DetalleDeEmpresa(ctx context.Context, identificador string) (DetalleEmpresaDePlataforma, bool, error)
	ExisteCodigoDeEmpresa(ctx context.Context, codigo string) (bool, error)
}
