package escudo

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	ErrNombreDeUsuarioInvalido = errors.New("el nombre de usuario debe tener de 3 a 40 caracteres: minusculas, numeros, punto, guion o guion bajo")
	ErrCorreoInvalido          = errors.New("el correo electronico no tiene un formato valido")
	ErrCodigoDeEmpresaInvalido = errors.New("el codigo de empresa debe tener de 2 a 12 letras mayusculas o numeros")
	ErrColorInvalido           = errors.New("el color debe tener formato #RRGGBB")
	ErrTextoDemasiadoLargo     = errors.New("el texto supera la longitud maxima permitida")
	ErrTextoConControl         = errors.New("el texto contiene caracteres de control no permitidos")
	ErrZonaHorariaInvalida     = errors.New("la zona horaria no existe")
	ErrTelefonoInvalido        = errors.New("el telefono solo admite numeros, espacios y los signos + - ( )")
)

var (
	patronNombreDeUsuario = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{2,39}$`)
	patronCorreo          = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]{2,}$`)
	patronCodigoEmpresa   = regexp.MustCompile(`^[A-Z0-9]{2,12}$`)
	patronColorHex        = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	patronTelefono        = regexp.MustCompile(`^[0-9+()\-\s]{6,20}$`)
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

type CampoDeTexto struct {
	Nombre string
	Valor  *string
	Maximo int
}

func Campo(nombre string, valor *string, maximo int) CampoDeTexto {
	return CampoDeTexto{Nombre: nombre, Valor: valor, Maximo: maximo}
}

func ValidarTextos(campos ...CampoDeTexto) error {
	for _, campo := range campos {
		limpio, err := TextoSeguro(*campo.Valor, campo.Maximo)
		if err != nil {
			if errors.Is(err, ErrTextoDemasiadoLargo) {
				return fmt.Errorf("el campo %s no puede pasar de %d caracteres", campo.Nombre, campo.Maximo)
			}
			return fmt.Errorf("el campo %s tiene caracteres no permitidos", campo.Nombre)
		}
		*campo.Valor = limpio
	}
	return nil
}

func ValidarZonaHoraria(valor string) error {
	if valor == "" {
		return nil
	}
	if _, err := time.LoadLocation(valor); err != nil {
		return ErrZonaHorariaInvalida
	}
	return nil
}

func ValidarTelefono(valor string) error {
	if valor == "" {
		return nil
	}
	if !patronTelefono.MatchString(valor) {
		return ErrTelefonoInvalido
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
