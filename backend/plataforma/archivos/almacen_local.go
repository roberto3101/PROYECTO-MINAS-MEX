package archivos

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
)

type AlmacenLocal struct {
	directorioRaiz string
}

func NuevoAlmacenLocal(directorioRaiz string) AlmacenLocal {
	return AlmacenLocal{directorioRaiz: directorioRaiz}
}

func (almacen AlmacenLocal) GuardarLogo(identificadorEmpresa, formato string, datos []byte) (string, error) {
	carpeta := filepath.Join(almacen.directorioRaiz, "logos")
	if err := os.MkdirAll(carpeta, 0o755); err != nil {
		return "", err
	}
	sello := make([]byte, 6)
	if _, err := rand.Read(sello); err != nil {
		return "", err
	}
	nombre := identificadorEmpresa + "-" + hex.EncodeToString(sello) + "." + extensionDe(formato)
	if err := os.WriteFile(filepath.Join(carpeta, nombre), datos, 0o644); err != nil {
		return "", err
	}
	return "/archivos/logos/" + nombre, nil
}

func extensionDe(formato string) string {
	if formato == "jpeg" {
		return "jpg"
	}
	return formato
}
