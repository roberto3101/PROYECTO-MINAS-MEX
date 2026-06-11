package dominio

import (
	"strings"

	"minas/compartido/identificador"
)

type Equipo struct {
	id              identificador.Identificador
	idEmpresa       identificador.Identificador
	idMina          identificador.Identificador
	idTipoEquipo    identificador.Identificador
	idModuloTrabajo identificador.Identificador
	codigo          string
	fabricante      string
}

func DarDeAltaEquipo(idEmpresa, idMina, idTipoEquipo, idModuloTrabajo identificador.Identificador, codigo, fabricante string) (Equipo, error) {
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
	return Equipo{
		id:              identificador.Nuevo(),
		idEmpresa:       idEmpresa,
		idMina:          idMina,
		idTipoEquipo:    idTipoEquipo,
		idModuloTrabajo: idModuloTrabajo,
		codigo:          codigoLimpio,
		fabricante:      strings.TrimSpace(fabricante),
	}, nil
}

func (equipo Equipo) Identificador() identificador.Identificador  { return equipo.id }
func (equipo Equipo) Empresa() identificador.Identificador        { return equipo.idEmpresa }
func (equipo Equipo) Mina() identificador.Identificador           { return equipo.idMina }
func (equipo Equipo) TipoEquipo() identificador.Identificador     { return equipo.idTipoEquipo }
func (equipo Equipo) ModuloTrabajo() identificador.Identificador  { return equipo.idModuloTrabajo }
func (equipo Equipo) Codigo() string                              { return equipo.codigo }
func (equipo Equipo) Fabricante() string                          { return equipo.fabricante }
