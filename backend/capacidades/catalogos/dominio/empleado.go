package dominio

import (
	"strings"

	"minas/compartido/identificador"
)

const (
	EmpleadoActivo   = "ACTIVO"
	EmpleadoInactivo = "INACTIVO"
)

var EstadosDeEmpleadoValidos = []string{EmpleadoActivo, EmpleadoInactivo}

type Empleado struct {
	id             identificador.Identificador
	idEmpresa      identificador.Identificador
	idMina         identificador.Identificador
	numeroNomina   string
	nombreCompleto string
	idDepartamento *identificador.Identificador
	idPuesto       *identificador.Identificador
	idActividad    *identificador.Identificador
	centroCosto    string
	gerenteACargo  string
	grupo          string
}

func ContratarEmpleado(idEmpresa, idMina identificador.Identificador, numeroNomina, nombreCompleto string, idDepartamento, idPuesto, idActividad *identificador.Identificador, centroCosto, gerenteACargo, grupo string) (Empleado, error) {
	if idMina.EsVacio() {
		return Empleado{}, ErrMinaObligatoria
	}
	nomina := strings.TrimSpace(numeroNomina)
	if nomina == "" {
		return Empleado{}, ErrNumeroNominaObligatorio
	}
	nombre := strings.ToUpper(strings.TrimSpace(nombreCompleto))
	if nombre == "" {
		return Empleado{}, ErrNombreObligatorio
	}
	return Empleado{
		id:             identificador.Nuevo(),
		idEmpresa:      idEmpresa,
		idMina:         idMina,
		numeroNomina:   nomina,
		nombreCompleto: nombre,
		idDepartamento: idDepartamento,
		idPuesto:       idPuesto,
		idActividad:    idActividad,
		centroCosto:    strings.TrimSpace(centroCosto),
		gerenteACargo:  strings.TrimSpace(gerenteACargo),
		grupo:          strings.TrimSpace(grupo),
	}, nil
}

func (empleado Empleado) Identificador() identificador.Identificador { return empleado.id }
func (empleado Empleado) Empresa() identificador.Identificador       { return empleado.idEmpresa }
func (empleado Empleado) Mina() identificador.Identificador          { return empleado.idMina }
func (empleado Empleado) NumeroNomina() string                       { return empleado.numeroNomina }
func (empleado Empleado) NombreCompleto() string                     { return empleado.nombreCompleto }
func (empleado Empleado) Departamento() *identificador.Identificador { return empleado.idDepartamento }
func (empleado Empleado) Puesto() *identificador.Identificador       { return empleado.idPuesto }
func (empleado Empleado) Actividad() *identificador.Identificador    { return empleado.idActividad }
func (empleado Empleado) CentroCosto() string                        { return empleado.centroCosto }
func (empleado Empleado) GerenteACargo() string                      { return empleado.gerenteACargo }
func (empleado Empleado) Grupo() string                              { return empleado.grupo }
