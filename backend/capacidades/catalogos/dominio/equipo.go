package dominio

import (
	"strings"
	"time"

	"minas/compartido/identificador"
)

const (
	EquipoOperativo       = "OPERATIVO"
	EquipoEnMantenimiento = "MANTENIMIENTO"
	EquipoDeBaja          = "BAJA"
)

var EstadosDeEquipoValidos = []string{EquipoOperativo, EquipoEnMantenimiento, EquipoDeBaja}

const (
	anioFabricacionMinimo = 1900
	anioFabricacionMaximo = 2100
)

type Equipo struct {
	id                identificador.Identificador
	idEmpresa         identificador.Identificador
	idMina            identificador.Identificador
	idTipoEquipo      identificador.Identificador
	idModuloTrabajo   identificador.Identificador
	codigo            string
	descripcion       string
	modeloPerforadora string
	capacidadLongitud *float64
	fabricante        string
	modelo            string
	numeroSerie       string
	anioFabricacion   *int
	fechaIngresoMina  *time.Time
	estado            string
}

func DarDeAltaEquipo(idEmpresa, idMina, idTipoEquipo, idModuloTrabajo identificador.Identificador, codigo, descripcion, modeloPerforadora, fabricante, modelo, numeroSerie string, capacidadLongitud *float64, anioFabricacion *int, fechaIngresoMina *time.Time) (Equipo, error) {
	if idMina.EsVacio() {
		return Equipo{}, ErrMinaObligatoria
	}
	if idTipoEquipo.EsVacio() {
		return Equipo{}, ErrTipoDeEquipoObligatorio
	}
	if idModuloTrabajo.EsVacio() {
		return Equipo{}, ErrModuloTrabajoObligatorio
	}
	codigoLimpio := strings.ToUpper(strings.TrimSpace(codigo))
	if codigoLimpio == "" {
		return Equipo{}, ErrCodigoObligatorio
	}
	if err := validarPositivos(capacidadLongitud); err != nil {
		return Equipo{}, err
	}
	if anioFabricacion != nil && (*anioFabricacion < anioFabricacionMinimo || *anioFabricacion > anioFabricacionMaximo) {
		return Equipo{}, ErrValorFueraDeRango
	}
	return Equipo{
		id:                identificador.Nuevo(),
		idEmpresa:         idEmpresa,
		idMina:            idMina,
		idTipoEquipo:      idTipoEquipo,
		idModuloTrabajo:   idModuloTrabajo,
		codigo:            codigoLimpio,
		descripcion:       strings.TrimSpace(descripcion),
		modeloPerforadora: strings.TrimSpace(modeloPerforadora),
		capacidadLongitud: capacidadLongitud,
		fabricante:        strings.TrimSpace(fabricante),
		modelo:            strings.TrimSpace(modelo),
		numeroSerie:       strings.TrimSpace(numeroSerie),
		anioFabricacion:   anioFabricacion,
		fechaIngresoMina:  fechaIngresoMina,
		estado:            EquipoOperativo,
	}, nil
}

func (equipo Equipo) Identificador() identificador.Identificador { return equipo.id }
func (equipo Equipo) Empresa() identificador.Identificador       { return equipo.idEmpresa }
func (equipo Equipo) Mina() identificador.Identificador          { return equipo.idMina }
func (equipo Equipo) TipoEquipo() identificador.Identificador    { return equipo.idTipoEquipo }
func (equipo Equipo) ModuloTrabajo() identificador.Identificador { return equipo.idModuloTrabajo }
func (equipo Equipo) Codigo() string                             { return equipo.codigo }
func (equipo Equipo) Descripcion() string                        { return equipo.descripcion }
func (equipo Equipo) ModeloPerforadora() string                  { return equipo.modeloPerforadora }
func (equipo Equipo) CapacidadLongitud() *float64                { return equipo.capacidadLongitud }
func (equipo Equipo) Fabricante() string                         { return equipo.fabricante }
func (equipo Equipo) Modelo() string                             { return equipo.modelo }
func (equipo Equipo) NumeroSerie() string                        { return equipo.numeroSerie }
func (equipo Equipo) AnioFabricacion() *int                      { return equipo.anioFabricacion }
func (equipo Equipo) FechaIngresoMina() *time.Time               { return equipo.fechaIngresoMina }
func (equipo Equipo) Estado() string                             { return equipo.estado }
