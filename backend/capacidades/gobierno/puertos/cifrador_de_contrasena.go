package puertos

type CifradorDeContrasena interface {
	Cifrar(textoPlano string) (string, error)
	Verificar(textoPlano, cifrado string) bool
}
