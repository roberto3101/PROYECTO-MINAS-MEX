package dominio

import (
	"strings"

	"minas/compartido/identificador"
)

type EstadoRol string

const (
	RolActivo   EstadoRol = "ACTIVO"
	RolInactivo EstadoRol = "INACTIVO"
)

type Rol struct {
	id          identificador.Identificador
	idEmpresa   identificador.Identificador
	codigo      string
	descripcion string
	esDeSistema bool
	permisos    []identificador.Identificador
	estado      EstadoRol
}

func CrearRolPropio(idEmpresa identificador.Identificador, codigo, descripcion string) (Rol, error) {
	codigoLimpio := strings.TrimSpace(codigo)
	if codigoLimpio == "" {
		return Rol{}, ErrCodigoObligatorio
	}
	descripcionLimpia := strings.TrimSpace(descripcion)
	if descripcionLimpia == "" {
		return Rol{}, ErrNombreObligatorio
	}
	return Rol{
		id:          identificador.Nuevo(),
		idEmpresa:   idEmpresa,
		codigo:      codigoLimpio,
		descripcion: descripcionLimpia,
		esDeSistema: false,
		estado:      RolActivo,
	}, nil
}

func ReconstruirRol(id, idEmpresa identificador.Identificador, codigo, descripcion string, esDeSistema bool, permisos []identificador.Identificador, estado EstadoRol) Rol {
	return Rol{id: id, idEmpresa: idEmpresa, codigo: codigo, descripcion: descripcion, esDeSistema: esDeSistema, permisos: permisos, estado: estado}
}

func (rol *Rol) Renombrar(descripcion string) error {
	if rol.esDeSistema {
		return ErrRolDeSistemaProtegido
	}
	limpia := strings.TrimSpace(descripcion)
	if limpia == "" {
		return ErrNombreObligatorio
	}
	rol.descripcion = limpia
	return nil
}

func (rol *Rol) ConcederPermiso(idPermiso identificador.Identificador) error {
	if rol.esDeSistema {
		return ErrRolDeSistemaProtegido
	}
	if rol.tienePermiso(idPermiso) {
		return ErrPermisoYaConcedido
	}
	rol.permisos = append(rol.permisos, idPermiso)
	return nil
}

func (rol *Rol) RevocarPermiso(idPermiso identificador.Identificador) error {
	if rol.esDeSistema {
		return ErrRolDeSistemaProtegido
	}
	for indice, concedido := range rol.permisos {
		if concedido.Igual(idPermiso) {
			rol.permisos = append(rol.permisos[:indice], rol.permisos[indice+1:]...)
			return nil
		}
	}
	return ErrPermisoNoConcedido
}

func (rol Rol) tienePermiso(idPermiso identificador.Identificador) bool {
	for _, concedido := range rol.permisos {
		if concedido.Igual(idPermiso) {
			return true
		}
	}
	return false
}

func (rol Rol) Identificador() identificador.Identificador { return rol.id }
func (rol Rol) Empresa() identificador.Identificador       { return rol.idEmpresa }
func (rol Rol) Codigo() string                             { return rol.codigo }
func (rol Rol) Descripcion() string                        { return rol.descripcion }
func (rol Rol) EsDeSistema() bool                          { return rol.esDeSistema }
func (rol Rol) Permisos() []identificador.Identificador    { return rol.permisos }
func (rol Rol) Estado() EstadoRol                          { return rol.estado }
