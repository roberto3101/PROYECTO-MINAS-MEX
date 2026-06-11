package identificador

import (
	"crypto/rand"
	"errors"
	"fmt"
	"regexp"
)

type Identificador struct {
	valor string
}

var ErrIdentificadorInvalido = errors.New("identificador invalido")

var patronUuid = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func Nuevo() Identificador {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	return Identificador{valor: fmt.Sprintf("%x-%x-%x-%x-%x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])}
}

func Desde(texto string) (Identificador, error) {
	if !patronUuid.MatchString(texto) {
		return Identificador{}, ErrIdentificadorInvalido
	}
	return Identificador{valor: texto}, nil
}

func DesdeOpcional(texto string) (*Identificador, error) {
	if texto == "" {
		return nil, nil
	}
	resuelto, err := Desde(texto)
	if err != nil {
		return nil, err
	}
	return &resuelto, nil
}

func (identificador Identificador) Texto() string {
	return identificador.valor
}

func (identificador Identificador) EsVacio() bool {
	return identificador.valor == ""
}

func (identificador Identificador) Igual(otro Identificador) bool {
	return identificador.valor == otro.valor
}
