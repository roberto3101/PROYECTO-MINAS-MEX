package dominio

import (
	"strings"

	"minas/compartido/identificador"
)

const (
	ObraActiva     = "ACTIVA"
	ObraSuspendida = "SUSPENDIDA"
	ObraConcluida  = "CONCLUIDA"
)

var EstadosDeObraValidos = []string{ObraActiva, ObraSuspendida, ObraConcluida}

type Obra struct {
	id            identificador.Identificador
	idEmpresa     identificador.Identificador
	idMina        identificador.Identificador
	idTipoObra    *identificador.Identificador
	codigo        string
	nombre        string
	ubicacion     string
	esPrioritaria bool
}

func CrearObra(idEmpresa, idMina identificador.Identificador, idTipoObra *identificador.Identificador,
	codigo, nombre, ubicacion string, esPrioritaria bool) (Obra, error) {
	codigoLimpio := strings.TrimSpace(codigo)
	if codigoLimpio == "" {
		return Obra{}, ErrCodigoObligatorio
	}
	return Obra{
		id:            identificador.Nuevo(),
		idEmpresa:     idEmpresa,
		idMina:        idMina,
		idTipoObra:    idTipoObra,
		codigo:        codigoLimpio,
		nombre:        strings.TrimSpace(nombre),
		ubicacion:     strings.TrimSpace(ubicacion),
		esPrioritaria: esPrioritaria,
	}, nil
}

func (obra Obra) Identificador() identificador.Identificador { return obra.id }
func (obra Obra) Empresa() identificador.Identificador       { return obra.idEmpresa }
func (obra Obra) Mina() identificador.Identificador          { return obra.idMina }
func (obra Obra) TipoDeObra() *identificador.Identificador   { return obra.idTipoObra }
func (obra Obra) Codigo() string                             { return obra.codigo }
func (obra Obra) Nombre() string                             { return obra.nombre }
func (obra Obra) Ubicacion() string                          { return obra.ubicacion }
func (obra Obra) EsPrioritaria() bool                        { return obra.esPrioritaria }
