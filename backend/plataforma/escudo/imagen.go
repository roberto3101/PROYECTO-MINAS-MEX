package escudo

import (
	"bytes"
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"
)

const (
	tamanoMaximoDeLogo = 2 << 20
	ladoMaximoDeLogo   = 2048
)

var (
	ErrImagenDemasiadoGrande = errors.New("la imagen supera el tamano maximo de 2 MB")
	ErrImagenNoPermitida     = errors.New("solo se aceptan imagenes PNG, JPEG o WebP reales")
	ErrImagenCorrupta        = errors.New("el archivo no es una imagen valida o esta dañado")
	ErrImagenDesmedida       = errors.New("la imagen no debe exceder 2048 x 2048 pixeles")
)

var firmasDeImagen = []struct {
	formato string
	magia   []byte
}{
	{"png", []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}},
	{"jpeg", []byte{0xFF, 0xD8, 0xFF}},
	{"webp", []byte("RIFF")},
}

func ValidarImagenDeLogo(datos []byte) (string, error) {
	if len(datos) > tamanoMaximoDeLogo {
		return "", ErrImagenDemasiadoGrande
	}
	formato := formatoPorMagia(datos)
	if formato == "" {
		return "", ErrImagenNoPermitida
	}
	configuracion, formatoDecodificado, err := image.DecodeConfig(bytes.NewReader(datos))
	if err != nil || formatoDecodificado != formato {
		return "", ErrImagenCorrupta
	}
	if configuracion.Width > ladoMaximoDeLogo || configuracion.Height > ladoMaximoDeLogo ||
		configuracion.Width < 1 || configuracion.Height < 1 {
		return "", ErrImagenDesmedida
	}
	return formato, nil
}

func formatoPorMagia(datos []byte) string {
	for _, firma := range firmasDeImagen {
		if bytes.HasPrefix(datos, firma.magia) {
			if firma.formato == "webp" {
				if len(datos) < 12 || !bytes.Equal(datos[8:12], []byte("WEBP")) {
					continue
				}
			}
			return firma.formato
		}
	}
	return ""
}
