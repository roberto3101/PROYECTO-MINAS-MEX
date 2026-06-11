package dominio

import (
	"strings"

	"minas/compartido/identificador"
)

type Empleado struct {
	id             identificador.Identificador
	idEmpresa      identificador.Identificador
	idMina         identificador.Identificador
	numeroNomina   string
	nombreCompleto string
}

func ContratarEmpleado(idEmpresa, idMina identificador.Identificador, numeroNomina, nombreCompleto string) (Empleado, error) {
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
	}, nil
}

func (empleado Empleado) Identificador() identificador.Identificador { return empleado.id }
func (empleado Empleado) Empresa() identificador.Identificador       { return empleado.idEmpresa }
func (empleado Empleado) Mina() identificador.Identificador          { return empleado.idMina }
func (empleado Empleado) NumeroNomina() string                       { return empleado.numeroNomina }
func (empleado Empleado) NombreCompleto() string                     { return empleado.nombreCompleto }
