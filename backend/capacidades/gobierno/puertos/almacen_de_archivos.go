package puertos

type AlmacenDeArchivos interface {
	GuardarLogo(identificadorEmpresa, formato string, datos []byte) (rutaPublica string, err error)
}
