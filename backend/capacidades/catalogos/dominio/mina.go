package dominio

import (
	"strings"

	"minas/compartido/identificador"
)

type Mina struct {
	id        identificador.Identificador
	idEmpresa identificador.Identificador
	nombre    string
	area      string
}

func CrearMina(idEmpresa identificador.Identificador, nombre, area string) (Mina, error) {
	nombreLimpio := strings.TrimSpace(nombre)
	if nombreLimpio == "" {
		return Mina{}, ErrNombreObligatorio
	}
	return Mina{
		id:        identificador.Nuevo(),
		idEmpresa: idEmpresa,
		nombre:    nombreLimpio,
		area:      strings.TrimSpace(area),
	}, nil
}

func (mina Mina) Identificador() identificador.Identificador { return mina.id }
func (mina Mina) Empresa() identificador.Identificador       { return mina.idEmpresa }
func (mina Mina) Nombre() string                             { return mina.nombre }
func (mina Mina) Area() string                               { return mina.area }
