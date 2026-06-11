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
func (branding Branding) Moneda() string        { return branding.moneda }

type PerfilDeContacto struct {
	IdentificacionFiscal string
	CorreoContacto       string
	Telefono             string
}

type Empresa struct {
	id          identificador.Identificador
	codigo      string
	razonSocial string
	perfil      PerfilDeContacto
	branding    Branding
	estado      EstadoEmpresa
}

func CrearEmpresa(codigo, razonSocial string, perfil PerfilDeContacto, branding Branding) (Empresa, error) {
	codigoLimpio := strings.ToUpper(strings.TrimSpace(codigo))
	if codigoLimpio == "" {
		return Empresa{}, ErrCodigoDeEmpresaObligatorio
	}
	razonLimpia := strings.TrimSpace(razonSocial)
	if razonLimpia == "" {
		return Empresa{}, ErrRazonSocialObligatoria
	}
	return Empresa{
		id:          identificador.Nuevo(),
		codigo:      codigoLimpio,
		razonSocial: razonLimpia,
		perfil:      perfil,
		branding:    branding,
		estado:      EmpresaActiva,
	}, nil
}

func ReconstruirEmpresa(id identificador.Identificador, codigo, razonSocial string, perfil PerfilDeContacto, branding Branding, estado EstadoEmpresa) Empresa {
	return Empresa{id: id, codigo: codigo, razonSocial: razonSocial, perfil: perfil, branding: branding, estado: estado}
}

func (empresa *Empresa) AplicarBranding(branding Branding) {
	empresa.branding = branding
}

func (empresa *Empresa) ActualizarPerfil(perfil PerfilDeContacto) {
	empresa.perfil = perfil
}

func (empresa *Empresa) DefinirLogo(ruta string) {
	empresa.branding.logoUrl = strings.TrimSpace(ruta)
}

func (empresa *Empresa) Activar()    { empresa.estado = EmpresaActiva }
func (empresa *Empresa) Desactivar() { empresa.estado = EmpresaInactiva }

func (empresa Empresa) Identificador() identificador.Identificador { return empresa.id }
func (empresa Empresa) Codigo() string                             { return empresa.codigo }
func (empresa Empresa) RazonSocial() string                        { return empresa.razonSocial }
func (empresa Empresa) Perfil() PerfilDeContacto                   { return empresa.perfil }
func (empresa Empresa) Branding() Branding                         { return empresa.branding }
func (empresa Empresa) Estado() EstadoEmpresa                      { return empresa.estado }
func (empresa Empresa) EstaActiva() bool                           { return empresa.estado == EmpresaActiva }
