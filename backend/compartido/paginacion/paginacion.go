package paginacion

import (
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
)

var ErrCursorInvalido = errors.New("el cursor de paginacion no es valido")

const (
	limitePorDefecto = 20
	limiteMaximo     = 100
)

func LimiteSeguro(texto string) int {
	valor, err := strconv.Atoi(texto)
	if err != nil || valor < 1 {
		return limitePorDefecto
	}
	if valor > limiteMaximo {
		return limiteMaximo
	}
	return valor
}

func CodificarCursor(ordenamiento, identificador string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(ordenamiento + "|" + identificador))
}

func DecodificarCursor(cursor string) (ordenamiento, identificador string, err error) {
	if cursor == "" {
		return "", "", nil
	}
	crudo, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return "", "", ErrCursorInvalido
	}
	partes := strings.SplitN(string(crudo), "|", 2)
	if len(partes) != 2 {
		return "", "", ErrCursorInvalido
	}
	return partes[0], partes[1], nil
}
