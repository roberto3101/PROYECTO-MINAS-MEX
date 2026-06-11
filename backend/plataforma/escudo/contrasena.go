package escudo

import (
	"errors"
	"unicode"
)

var ErrContrasenaDebil = errors.New("la contrasena debe tener al menos 10 caracteres e incluir mayuscula, minuscula y numero")

func ValidarContrasena(valor string) error {
	if len(valor) < 10 || len(valor) > 128 {
		return ErrContrasenaDebil
	}
	var tieneMayuscula, tieneMinuscula, tieneNumero bool
	for _, caracter := range valor {
		switch {
		case unicode.IsUpper(caracter):
			tieneMayuscula = true
		case unicode.IsLower(caracter):
			tieneMinuscula = true
		case unicode.IsDigit(caracter):
			tieneNumero = true
		}
	}
	if !tieneMayuscula || !tieneMinuscula || !tieneNumero {
		return ErrContrasenaDebil
	}
	return nil
}
