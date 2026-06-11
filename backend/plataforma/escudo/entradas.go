package escudo

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrNombreDeUsuarioInvalido = errors.New("el nombre de usuario debe tener de 3 a 40 caracteres: minusculas, numeros, punto, guion o guion bajo")
	ErrCorreoInvalido          = errors.New("el correo electronico no tiene un formato valido")
	ErrCodigoDeEmpresaInvalido = errors.New("el codigo de empresa debe tener de 2 a 12 letras mayusculas o numeros")
	ErrColorInvalido           = errors.New("el color debe tener formato #RRGGBB")
	ErrTextoDemasiadoLargo     = errors.New("el texto supera la longitud maxima permitida")
	ErrTextoConControl         = errors.New("el texto contiene caracteres de control no permitidos")
)

var (
	patronNombreDeUsuario = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{2,39}$`)
	patronCorreo          = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]{2,}$`)
	patronCodigoEmpresa   = regexp.MustCompile(`^[A-Z0-9]{2,12}$`)
	patronColorHex        = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
)

func ValidarNombreDeUsuario(valor string) error {
	if !patronNombreDeUsuario.MatchString(valor) {
		return ErrNombreDeUsuarioInvalido
	}
	return nil
}

func ValidarCorreo(valor string) error {
	if valor == "" {
		return nil
	}
	if len(valor) > 254 || !patronCorreo.MatchString(valor) {
		return ErrCorreoInvalido
	}
	return nil
}

func ValidarCodigoDeEmpresa(valor string) error {
	if !patronCodigoEmpresa.MatchString(valor) {
		return ErrCodigoDeEmpresaInvalido
	}
	return nil
}

func ValidarColor(valor string) error {
	if valor == "" {
		return nil
	}
	if !patronColorHex.MatchString(valor) {
		return ErrColorInvalido
	}
	return nil
}

func TextoSeguro(valor string, maximo int) (string, error) {
	limpio := strings.TrimSpace(valor)
	if len(limpio) > maximo {
		return "", ErrTextoDemasiadoLargo
	}
	for _, caracter := range limpio {
		if caracter < 32 && caracter != '\n' && caracter != '\t' {
			return "", ErrTextoConControl
		}
	}
	return limpio, nil
}
