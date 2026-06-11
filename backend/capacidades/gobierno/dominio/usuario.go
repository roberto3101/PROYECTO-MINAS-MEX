package dominio

import (
	"strings"

	"minas/compartido/identificador"
)

type EstadoUsuario string

const (
	UsuarioActivo   EstadoUsuario = "ACTIVO"
	UsuarioInactivo EstadoUsuario = "INACTIVO"
)

type Usuario struct {
	id                identificador.Identificador
	idEmpresa         identificador.Identificador
	nombreCorto       string
	nombre            string
	correo            string
	contrasenaCifrada string
	idEmpleado        *identificador.Identificador
	estado            EstadoUsuario
}

func RegistrarUsuario(idEmpresa identificador.Identificador, nombreCorto, nombre string) (Usuario, error) {
	corto := strings.TrimSpace(nombreCorto)
	if corto == "" {
		return Usuario{}, ErrCodigoObligatorio
	}
	completo := strings.TrimSpace(nombre)
	if completo == "" {
		return Usuario{}, ErrNombreObligatorio
	}
	return Usuario{
		id:          identificador.Nuevo(),
		idEmpresa:   idEmpresa,
		nombreCorto: corto,
		nombre:      completo,
		estado:      UsuarioActivo,
	}, nil
}

func ReconstruirUsuario(id, idEmpresa identificador.Identificador, nombreCorto, nombre, correo string, idEmpleado *identificador.Identificador, estado EstadoUsuario) Usuario {
	return Usuario{id: id, idEmpresa: idEmpresa, nombreCorto: nombreCorto, nombre: nombre, correo: correo, idEmpleado: idEmpleado, estado: estado}
}

func (usuario *Usuario) DefinirCorreo(correo Correo) {
	usuario.correo = correo.Texto()
}

func (usuario *Usuario) DefinirContrasenaCifrada(cifrada string) {
	usuario.contrasenaCifrada = cifrada
}

func (usuario *Usuario) VincularConEmpleado(idEmpleado identificador.Identificador) {
	usuario.idEmpleado = &idEmpleado
}

func (usuario *Usuario) Desactivar() error {
	if usuario.estado == UsuarioInactivo {
		return ErrUsuarioYaInactivo
	}
	usuario.estado = UsuarioInactivo
	return nil
}

func (usuario Usuario) Identificador() identificador.Identificador { return usuario.id }
func (usuario Usuario) Empresa() identificador.Identificador       { return usuario.idEmpresa }
func (usuario Usuario) NombreCorto() string                        { return usuario.nombreCorto }
func (usuario Usuario) Nombre() string                             { return usuario.nombre }
func (usuario Usuario) Correo() string                             { return usuario.correo }
func (usuario Usuario) ContrasenaCifrada() string                  { return usuario.contrasenaCifrada }
func (usuario Usuario) EmpleadoVinculado() *identificador.Identificador { return usuario.idEmpleado }
func (usuario Usuario) Estado() EstadoUsuario                      { return usuario.estado }
func (usuario Usuario) EstaActivo() bool                           { return usuario.estado == UsuarioActivo }
