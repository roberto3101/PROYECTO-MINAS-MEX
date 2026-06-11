package dominio

import "minas/compartido/identificador"

type Permiso struct {
	id          identificador.Identificador
	codigo      string
	descripcion string
	modulo      string
}

func ReconstruirPermiso(id identificador.Identificador, codigo, descripcion, modulo string) Permiso {
	return Permiso{id: id, codigo: codigo, descripcion: descripcion, modulo: modulo}
}

func (permiso Permiso) Identificador() identificador.Identificador { return permiso.id }
func (permiso Permiso) Codigo() string                             { return permiso.codigo }
func (permiso Permiso) Descripcion() string                        { return permiso.descripcion }
func (permiso Permiso) Modulo() string                             { return permiso.modulo }
