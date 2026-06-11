package dominio

import (
	"regexp"
	"strings"
)

type Correo struct {
	valor string
}

var patronCorreo = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

func CorreoDesde(texto string) (Correo, error) {
	limpio := strings.TrimSpace(texto)
	if !patronCorreo.MatchString(limpio) {
		return Correo{}, ErrCorreoInvalido
	}
	return Correo{valor: limpio}, nil
}

func (correo Correo) Texto() string {
	return correo.valor
}

var patronColorHexadecimal = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

func esColorHexadecimalValido(color string) bool {
	return patronColorHexadecimal.MatchString(color)
}

var monedasSoportadas = map[string]bool{"USD": true, "MXN": true}

const zonaHorariaPorDefecto = "America/Mexico_City"
const monedaPorDefecto = "USD"
