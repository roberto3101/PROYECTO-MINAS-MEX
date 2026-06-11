package identidad

type Sesion struct {
	IdentificadorUsuario string
	IdentificadorEmpresa string
	NombreCorto          string
	Permisos             []string
}
