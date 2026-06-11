package dominio

import (
	"errors"
	"testing"

	"minas/compartido/identificador"
)

func TestRegistrarUsuarioExigeNombreCorto(t *testing.T) {
	_, err := RegistrarUsuario(identificador.Nuevo(), "   ", "Operador")
	if !errors.Is(err, ErrCodigoObligatorio) {
		t.Fatalf("se esperaba ErrCodigoObligatorio, se obtuvo %v", err)
	}
}

func TestRegistrarUsuarioExigeNombre(t *testing.T) {
	_, err := RegistrarUsuario(identificador.Nuevo(), "op", "")
	if !errors.Is(err, ErrNombreObligatorio) {
		t.Fatalf("se esperaba ErrNombreObligatorio, se obtuvo %v", err)
	}
}

func TestDesactivarUsuarioDosVecesFalla(t *testing.T) {
	usuario, err := RegistrarUsuario(identificador.Nuevo(), "op", "Operador")
	if err != nil {
		t.Fatalf("registro inesperadamente fallido: %v", err)
	}
	if err := usuario.Desactivar(); err != nil {
		t.Fatalf("primera desactivacion fallida: %v", err)
	}
	if err := usuario.Desactivar(); !errors.Is(err, ErrUsuarioYaInactivo) {
		t.Fatalf("se esperaba ErrUsuarioYaInactivo, se obtuvo %v", err)
	}
}

func TestBrandingRechazaColorInvalido(t *testing.T) {
	_, err := ConfigurarBranding("", "naranja", "", "")
	if !errors.Is(err, ErrColorDeMarcaInvalido) {
		t.Fatalf("se esperaba ErrColorDeMarcaInvalido, se obtuvo %v", err)
	}
}

func TestBrandingAplicaValoresPorDefecto(t *testing.T) {
	branding, err := ConfigurarBranding("", "#0E7A4B", "", "")
	if err != nil {
		t.Fatalf("branding valido inesperadamente fallido: %v", err)
	}
	if branding.ZonaHoraria() != "America/Mexico_City" || branding.Moneda() != "USD" {
		t.Fatalf("valores por defecto incorrectos: %q %q", branding.ZonaHoraria(), branding.Moneda())
	}
}

func TestRolDeSistemaNoAceptaPermisos(t *testing.T) {
	rol := ReconstruirRol(identificador.Nuevo(), identificador.Nuevo(), "OPERADOR", "Operador", true, nil, RolActivo)
	if err := rol.ConcederPermiso(identificador.Nuevo()); !errors.Is(err, ErrRolDeSistemaProtegido) {
		t.Fatalf("se esperaba ErrRolDeSistemaProtegido, se obtuvo %v", err)
	}
}

func TestRolPropioRechazaPermisoDuplicado(t *testing.T) {
	rol, err := CrearRolPropio(identificador.Nuevo(), "SUPERVISOR_PATIO", "Supervisor de patio")
	if err != nil {
		t.Fatalf("creacion de rol fallida: %v", err)
	}
	permiso := identificador.Nuevo()
	if err := rol.ConcederPermiso(permiso); err != nil {
		t.Fatalf("primera concesion fallida: %v", err)
	}
	if err := rol.ConcederPermiso(permiso); !errors.Is(err, ErrPermisoYaConcedido) {
		t.Fatalf("se esperaba ErrPermisoYaConcedido, se obtuvo %v", err)
	}
}

func TestAsignacionNoSeRevocaDosVeces(t *testing.T) {
	asignacion := AsignarRol(identificador.Nuevo(), identificador.Nuevo(), identificador.Nuevo(), nil)
	if err := asignacion.Revocar(); err != nil {
		t.Fatalf("primera revocacion fallida: %v", err)
	}
	if err := asignacion.Revocar(); !errors.Is(err, ErrAsignacionYaRevocada) {
		t.Fatalf("se esperaba ErrAsignacionYaRevocada, se obtuvo %v", err)
	}
}

func TestCorreoRechazaFormatoInvalido(t *testing.T) {
	if _, err := CorreoDesde("sin-arroba"); !errors.Is(err, ErrCorreoInvalido) {
		t.Fatalf("se esperaba ErrCorreoInvalido, se obtuvo %v", err)
	}
}
