package identidad

import (
	"errors"
	"testing"
	"time"
)

func emisorDePrueba() EmisorDeToken {
	return NuevoEmisorDeToken("secreto-de-prueba", time.Hour)
}

func sesionDePrueba() Sesion {
	return Sesion{
		IdentificadorUsuario: "11111111-1111-1111-1111-111111111111",
		IdentificadorEmpresa: "22222222-2222-2222-2222-222222222222",
		NombreCorto:          "admin.mina",
		Permisos:             []string{"usuarios.crear", "roles.asignar"},
	}
}

func TestTokenIdaYVuelta(t *testing.T) {
	emisor := emisorDePrueba()
	ahora := time.Unix(1_700_000_000, 0)
	token, err := emisor.Emitir(sesionDePrueba(), ahora)
	if err != nil {
		t.Fatalf("emision fallida: %v", err)
	}
	sesion, err := emisor.Verificar(token, ahora.Add(time.Minute))
	if err != nil {
		t.Fatalf("verificacion fallida: %v", err)
	}
	if sesion.IdentificadorUsuario != sesionDePrueba().IdentificadorUsuario {
		t.Fatalf("usuario incorrecto: %q", sesion.IdentificadorUsuario)
	}
	if len(sesion.Permisos) != 2 {
		t.Fatalf("se esperaban 2 permisos, se obtuvieron %d", len(sesion.Permisos))
	}
}

func TestTokenExpirado(t *testing.T) {
	emisor := emisorDePrueba()
	ahora := time.Unix(1_700_000_000, 0)
	token, _ := emisor.Emitir(sesionDePrueba(), ahora)
	if _, err := emisor.Verificar(token, ahora.Add(2*time.Hour)); !errors.Is(err, ErrTokenExpirado) {
		t.Fatalf("se esperaba ErrTokenExpirado, se obtuvo %v", err)
	}
}

func TestTokenManipuladoEsInvalido(t *testing.T) {
	emisor := emisorDePrueba()
	ahora := time.Unix(1_700_000_000, 0)
	token, _ := emisor.Emitir(sesionDePrueba(), ahora)
	if _, err := emisor.Verificar(token+"x", ahora); !errors.Is(err, ErrTokenInvalido) {
		t.Fatalf("se esperaba ErrTokenInvalido, se obtuvo %v", err)
	}
}
