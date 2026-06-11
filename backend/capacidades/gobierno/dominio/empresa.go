package dominio

import (
	"strings"

	"minas/compartido/identificador"
)

type EstadoEmpresa string

const (
	EmpresaActiva   EstadoEmpresa = "ACTIVA"
	EmpresaInactiva EstadoEmpresa = "INACTIVA"
)

type Branding struct {
	logoUrl       string
	colorPrimario string
	zonaHoraria   string
	moneda        string
}

func ConfigurarBranding(logoUrl, colorPrimario, zonaHoraria, moneda string) (Branding, error) {
	if colorPrimario != "" && !esColorHexadecimalValido(colorPrimario) {
		return Branding{}, ErrColorDeMarcaInvalido
	}
	zonaElegida := strings.TrimSpace(zonaHoraria)
	if zonaElegida == "" {
		zonaElegida = zonaHorariaPorDefecto
	}
	monedaElegida := strings.TrimSpace(moneda)
	if monedaElegida == "" {
		monedaElegida = monedaPorDefecto
	}
	if !monedasSoportadas[monedaElegida] {
		return Branding{}, ErrMonedaNoSoportada
	}
	return Branding{logoUrl: strings.TrimSpace(logoUrl), colorPrimario: colorPrimario, zonaHoraria: zonaElegida, moneda: monedaElegida}, nil
}

func (branding Branding) LogoUrl() string       { return branding.logoUrl }
func (branding Branding) ColorPrimario() string { return branding.colorPrimario }
func (branding Branding) ZonaHoraria() string   { return branding.zonaHoraria }
func (branding Branding) Moneda() string         { return branding.moneda }

type Empresa struct {
	id          identificador.Identificador
	codigo      string
	razonSocial string
	branding    Branding
	estado      EstadoEmpresa
}

func ReconstruirEmpresa(id identificador.Identificador, codigo, razonSocial string, branding Branding, estado EstadoEmpresa) Empresa {
	return Empresa{id: id, codigo: codigo, razonSocial: razonSocial, branding: branding, estado: estado}
}

func (empresa *Empresa) AplicarBranding(branding Branding) {
	empresa.branding = branding
}

func (empresa Empresa) Identificador() identificador.Identificador { return empresa.id }
func (empresa Empresa) Codigo() string                             { return empresa.codigo }
func (empresa Empresa) RazonSocial() string                        { return empresa.razonSocial }
func (empresa Empresa) Branding() Branding                         { return empresa.branding }
func (empresa Empresa) Estado() EstadoEmpresa                      { return empresa.estado }
func (empresa Empresa) EstaActiva() bool                           { return empresa.estado == EmpresaActiva }
