package identidad

const (
	AmbitoEmpresa    = "empresa"
	AmbitoPlataforma = "plataforma"
)

type Sesion struct {
	IdentificadorUsuario string
	IdentificadorEmpresa string
	NombreCorto          string
	Ambito               string
	Permisos             []string
}
