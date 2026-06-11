package escudo

import (
	"bytes"
	"encoding/base64"
	"errors"
	"testing"
	"time"
)

func TestContrasenaExigePolitica(t *testing.T) {
	debiles := []string{"corta1A", "sinmayusculas123", "SINMINUSCULAS123", "SinNumerosAqui"}
	for _, contrasena := range debiles {
		if err := ValidarContrasena(contrasena); !errors.Is(err, ErrContrasenaDebil) {
			t.Fatalf("%q deberia ser debil", contrasena)
		}
	}
	if err := ValidarContrasena("Operador#2026"); err != nil {
		t.Fatalf("contrasena valida rechazada: %v", err)
	}
}

func TestNombreDeUsuario(t *testing.T) {
	if err := ValidarNombreDeUsuario("admin.mina"); err != nil {
		t.Fatalf("usuario valido rechazado: %v", err)
	}
	invalidos := []string{"ab", "Admin", "user name", "user;drop", string(make([]byte, 60))}
	for _, usuario := range invalidos {
		if ValidarNombreDeUsuario(usuario) == nil {
			t.Fatalf("%q deberia ser invalido", usuario)
		}
	}
}

func TestImagenDeLogoAceptaPngRealYRechazaDisfraces(t *testing.T) {
	pngDeUnPixel, _ := base64.StdEncoding.DecodeString(
		"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==")
	if formato, err := ValidarImagenDeLogo(pngDeUnPixel); err != nil || formato != "png" {
		t.Fatalf("png real rechazado: formato=%q err=%v", formato, err)
	}
	if _, err := ValidarImagenDeLogo([]byte("MZ\x90\x00programa.exe")); !errors.Is(err, ErrImagenNoPermitida) {
		t.Fatalf("ejecutable disfrazado deberia rechazarse, err=%v", err)
	}
	disfraz := append([]byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}, []byte("<script>alert(1)</script>")...)
	if _, err := ValidarImagenDeLogo(disfraz); !errors.Is(err, ErrImagenCorrupta) {
		t.Fatalf("cabecera png con contenido falso deberia rechazarse, err=%v", err)
	}
	if _, err := ValidarImagenDeLogo(bytes.Repeat([]byte{1}, 3<<20)); !errors.Is(err, ErrImagenDemasiadoGrande) {
		t.Fatal("imagen de 3MB deberia rechazarse")
	}
}

func TestLimitadorBloqueaTrasCincoFallosYEscala(t *testing.T) {
	limitador := NuevoLimitadorDeIntentos()
	ahora := time.Now()
	for intento := 0; intento < 5; intento++ {
		if _, permitido := limitador.Permitir("ip|usuario", ahora); !permitido {
			t.Fatalf("intento %d deberia permitirse", intento)
		}
		limitador.RegistrarFallo("ip|usuario", ahora)
	}
	espera, permitido := limitador.Permitir("ip|usuario", ahora)
	if permitido || espera < 59 {
		t.Fatalf("tras 5 fallos deberia bloquear ~60s, espera=%d permitido=%v", espera, permitido)
	}
	if _, permitido := limitador.Permitir("ip|usuario", ahora.Add(61*time.Second)); !permitido {
		t.Fatal("pasado el bloqueo deberia permitir")
	}
	limitador.RegistrarExito("ip|usuario")
	if _, permitido := limitador.Permitir("ip|usuario", ahora); !permitido {
		t.Fatal("tras exito el registro se limpia")
	}
}
