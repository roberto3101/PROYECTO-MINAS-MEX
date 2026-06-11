package dominio

import (
	"errors"
	"testing"

	"minas/compartido/identificador"
)

func TestCrearMinaExigeNombre(t *testing.T) {
	if _, err := CrearMina(identificador.Nuevo(), "   ", ""); !errors.Is(err, ErrNombreObligatorio) {
		t.Fatalf("se esperaba ErrNombreObligatorio, se obtuvo %v", err)
	}
}

func TestContratarEmpleadoNormalizaElNombre(t *testing.T) {
	empleado, err := ContratarEmpleado(identificador.Nuevo(), identificador.Nuevo(), "C3-001", "  juan perez  ")
	if err != nil {
		t.Fatalf("contratacion fallida: %v", err)
	}
	if empleado.NombreCompleto() != "JUAN PEREZ" {
		t.Fatalf("nombre no normalizado: %q", empleado.NombreCompleto())
	}
}

func TestContratarEmpleadoExigeMina(t *testing.T) {
	if _, err := ContratarEmpleado(identificador.Nuevo(), identificador.Identificador{}, "C3-001", "JUAN"); !errors.Is(err, ErrMinaObligatoria) {
		t.Fatalf("se esperaba ErrMinaObligatoria, se obtuvo %v", err)
	}
}

func TestDarDeAltaEquipoNormalizaElCodigo(t *testing.T) {
	equipo, err := DarDeAltaEquipo(identificador.Nuevo(), identificador.Nuevo(), identificador.Nuevo(), identificador.Nuevo(), "  c3-eq01 ", "Sandvik")
	if err != nil {
		t.Fatalf("alta fallida: %v", err)
	}
	if equipo.Codigo() != "C3-EQ01" {
		t.Fatalf("codigo no normalizado: %q", equipo.Codigo())
	}
}

func TestDarDeAltaEquipoExigeTipo(t *testing.T) {
	_, err := DarDeAltaEquipo(identificador.Nuevo(), identificador.Nuevo(), identificador.Identificador{}, identificador.Nuevo(), "EQ", "")
	if !errors.Is(err, ErrTipoDeEquipoObligatorio) {
		t.Fatalf("se esperaba ErrTipoDeEquipoObligatorio, se obtuvo %v", err)
	}
}
