package seguridad

import "golang.org/x/crypto/bcrypt"

type CifradorBcrypt struct {
	costo int
}

func NuevoCifradorBcrypt() CifradorBcrypt {
	return CifradorBcrypt{costo: bcrypt.DefaultCost}
}

func (cifrador CifradorBcrypt) Cifrar(textoPlano string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(textoPlano), cifrador.costo)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (cifrador CifradorBcrypt) Verificar(textoPlano, cifrado string) bool {
	return bcrypt.CompareHashAndPassword([]byte(cifrado), []byte(textoPlano)) == nil
}
