package identidad

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	ErrTokenInvalido = errors.New("token invalido")
	ErrTokenExpirado = errors.New("token expirado")
)

type EmisorDeToken struct {
	secreto  []byte
	duracion time.Duration
}

func NuevoEmisorDeToken(secreto string, duracion time.Duration) EmisorDeToken {
	return EmisorDeToken{secreto: []byte(secreto), duracion: duracion}
}

type cuerpoDeToken struct {
	Sujeto   string   `json:"sub"`
	Empresa  string   `json:"emp"`
	Nombre   string   `json:"nom"`
	Ambito   string   `json:"amb"`
	Permisos []string `json:"per"`
	Expira   int64    `json:"exp"`
}

func (emisor EmisorDeToken) Emitir(sesion Sesion, emitidoEn time.Time) (string, error) {
	cabecera := codificar([]byte(`{"alg":"HS256","typ":"JWT"}`))
	cuerpo := cuerpoDeToken{
		Sujeto:   sesion.IdentificadorUsuario,
		Empresa:  sesion.IdentificadorEmpresa,
		Nombre:   sesion.NombreCorto,
		Ambito:   sesion.Ambito,
		Permisos: sesion.Permisos,
		Expira:   emitidoEn.Add(emisor.duracion).Unix(),
	}
	cuerpoSerializado, err := json.Marshal(cuerpo)
	if err != nil {
		return "", err
	}
	contenido := cabecera + "." + codificar(cuerpoSerializado)
	return contenido + "." + emisor.firmar(contenido), nil
}

func (emisor EmisorDeToken) Verificar(token string, ahora time.Time) (Sesion, error) {
	partes := strings.Split(token, ".")
	if len(partes) != 3 {
		return Sesion{}, ErrTokenInvalido
	}
	contenido := partes[0] + "." + partes[1]
	if !hmac.Equal([]byte(partes[2]), []byte(emisor.firmar(contenido))) {
		return Sesion{}, ErrTokenInvalido
	}
	cuerpoSerializado, err := decodificar(partes[1])
	if err != nil {
		return Sesion{}, ErrTokenInvalido
	}
	var cuerpo cuerpoDeToken
	if err := json.Unmarshal(cuerpoSerializado, &cuerpo); err != nil {
		return Sesion{}, ErrTokenInvalido
	}
	if ahora.Unix() >= cuerpo.Expira {
		return Sesion{}, ErrTokenExpirado
	}
	return Sesion{
		IdentificadorUsuario: cuerpo.Sujeto,
		IdentificadorEmpresa: cuerpo.Empresa,
		NombreCorto:          cuerpo.Nombre,
		Ambito:               cuerpo.Ambito,
		Permisos:             cuerpo.Permisos,
	}, nil
}

func (emisor EmisorDeToken) firmar(contenido string) string {
	autenticador := hmac.New(sha256.New, emisor.secreto)
	autenticador.Write([]byte(contenido))
	return base64.RawURLEncoding.EncodeToString(autenticador.Sum(nil))
}

func codificar(bytes []byte) string {
	return base64.RawURLEncoding.EncodeToString(bytes)
}

func decodificar(texto string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(texto)
}
