package dominio

import "minas/compartido/identificador"

type AsignacionRol struct {
	id          identificador.Identificador
	idEmpresa   identificador.Identificador
	idUsuario   identificador.Identificador
	idRol       identificador.Identificador
	alcanceMina *identificador.Identificador
	revocada    bool
}

func AsignarRol(idEmpresa, idUsuario, idRol identificador.Identificador, alcanceMina *identificador.Identificador) AsignacionRol {
	return AsignacionRol{
		id:          identificador.Nuevo(),
		idEmpresa:   idEmpresa,
		idUsuario:   idUsuario,
		idRol:       idRol,
		alcanceMina: alcanceMina,
		revocada:    false,
	}
}

func ReconstruirAsignacion(id, idEmpresa, idUsuario, idRol identificador.Identificador, alcanceMina *identificador.Identificador, revocada bool) AsignacionRol {
	return AsignacionRol{id: id, idEmpresa: idEmpresa, idUsuario: idUsuario, idRol: idRol, alcanceMina: alcanceMina, revocada: revocada}
}

func (asignacion *AsignacionRol) Revocar() error {
	if asignacion.revocada {
		return ErrAsignacionYaRevocada
	}
	asignacion.revocada = true
	return nil
}

func (asignacion AsignacionRol) Identificador() identificador.Identificador { return asignacion.id }
func (asignacion AsignacionRol) Empresa() identificador.Identificador       { return asignacion.idEmpresa }
func (asignacion AsignacionRol) Usuario() identificador.Identificador       { return asignacion.idUsuario }
func (asignacion AsignacionRol) Rol() identificador.Identificador           { return asignacion.idRol }
func (asignacion AsignacionRol) AlcanceMina() *identificador.Identificador {
	return asignacion.alcanceMina
}
func (asignacion AsignacionRol) EstaVigente() bool { return !asignacion.revocada }
