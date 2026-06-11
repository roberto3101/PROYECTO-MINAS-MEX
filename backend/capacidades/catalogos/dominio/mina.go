package dominio

import (
	"strings"

	"minas/compartido/identificador"
)

const (
	MinaActiva   = "ACTIVA"
	MinaInactiva = "INACTIVA"
)

var EstadosDeMinaValidos = []string{MinaActiva, MinaInactiva}

type Mina struct {
	id               identificador.Identificador
	idEmpresa        identificador.Identificador
	nombre           string
	area             string
	niveles          string
	densidadPromedio *float64
	seccionObraLargo *float64
	seccionObraAncho *float64
	seccionVetaLargo *float64
	seccionVetaAncho *float64
}

func CrearMina(idEmpresa identificador.Identificador, nombre, area, niveles string, densidadPromedio, seccionObraLargo, seccionObraAncho, seccionVetaLargo, seccionVetaAncho *float64) (Mina, error) {
	nombreLimpio := strings.TrimSpace(nombre)
	if nombreLimpio == "" {
		return Mina{}, ErrNombreObligatorio
	}
	areaLimpia := strings.TrimSpace(area)
	if areaLimpia == "" {
		return Mina{}, ErrAreaObligatoria
	}
	if err := validarPositivos(densidadPromedio, seccionObraLargo, seccionObraAncho, seccionVetaLargo, seccionVetaAncho); err != nil {
		return Mina{}, err
	}
	return Mina{
		id:               identificador.Nuevo(),
		idEmpresa:        idEmpresa,
		nombre:           nombreLimpio,
		area:             areaLimpia,
		niveles:          strings.TrimSpace(niveles),
		densidadPromedio: densidadPromedio,
		seccionObraLargo: seccionObraLargo,
		seccionObraAncho: seccionObraAncho,
		seccionVetaLargo: seccionVetaLargo,
		seccionVetaAncho: seccionVetaAncho,
	}, nil
}

func validarPositivos(valores ...*float64) error {
	for _, valor := range valores {
		if valor != nil && *valor <= 0 {
			return ErrValorFueraDeRango
		}
	}
	return nil
}

func (mina Mina) Identificador() identificador.Identificador { return mina.id }
func (mina Mina) Empresa() identificador.Identificador       { return mina.idEmpresa }
func (mina Mina) Nombre() string                             { return mina.nombre }
func (mina Mina) Area() string                               { return mina.area }
func (mina Mina) Niveles() string                            { return mina.niveles }
func (mina Mina) DensidadPromedio() *float64                 { return mina.densidadPromedio }
func (mina Mina) SeccionObraLargo() *float64                 { return mina.seccionObraLargo }
func (mina Mina) SeccionObraAncho() *float64                 { return mina.seccionObraAncho }
func (mina Mina) SeccionVetaLargo() *float64                 { return mina.seccionVetaLargo }
func (mina Mina) SeccionVetaAncho() *float64                 { return mina.seccionVetaAncho }
