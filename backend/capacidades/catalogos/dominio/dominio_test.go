package dominio

import (
	"errors"
	"testing"

	"minas/compartido/identificador"
)

func TestCrearMinaExigeNombre(t *testing.T) {
	if _, err := CrearMina(identificador.Nuevo(), "   ", "Zona Norte", "", nil, nil, nil, nil, nil); !errors.Is(err, ErrNombreObligatorio) {
		t.Fatalf("se esperaba ErrNombreObligatorio, se obtuvo %v", err)
	}
}

func TestCrearMinaExigeArea(t *testing.T) {
	if _, err := CrearMina(identificador.Nuevo(), "San Rafael", "   ", "", nil, nil, nil, nil, nil); !errors.Is(err, ErrAreaObligatoria) {
		t.Fatalf("se esperaba ErrAreaObligatoria, se obtuvo %v", err)
	}
}

func TestCrearMinaRechazaNumericosNoPositivos(t *testing.T) {
	densidad := -2.5
	if _, err := CrearMina(identificador.Nuevo(), "San Rafael", "Zona Norte", "", &densidad, nil, nil, nil, nil); !errors.Is(err, ErrValorFueraDeRango) {
		t.Fatalf("se esperaba ErrValorFueraDeRango, se obtuvo %v", err)
	}
}

func TestContratarEmpleadoNormalizaElNombre(t *testing.T) {
	empleado, err := ContratarEmpleado(identificador.Nuevo(), identificador.Nuevo(), "C3-001", "  juan perez  ", nil, nil, nil, "", "", "")
	if err != nil {
		t.Fatalf("contratacion fallida: %v", err)
	}
	if empleado.NombreCompleto() != "JUAN PEREZ" {
		t.Fatalf("nombre no normalizado: %q", empleado.NombreCompleto())
	}
}

func TestContratarEmpleadoExigeMina(t *testing.T) {
	if _, err := ContratarEmpleado(identificador.Nuevo(), identificador.Identificador{}, "C3-001", "JUAN", nil, nil, nil, "", "", ""); !errors.Is(err, ErrMinaObligatoria) {
		t.Fatalf("se esperaba ErrMinaObligatoria, se obtuvo %v", err)
	}
}

func TestDarDeAltaEquipoNormalizaElCodigo(t *testing.T) {
	equipo, err := DarDeAltaEquipo(identificador.Nuevo(), identificador.Nuevo(), identificador.Nuevo(), identificador.Nuevo(), "  c3-eq01 ", "", "", "Sandvik", "", "", nil, nil, nil)
	if err != nil {
		t.Fatalf("alta fallida: %v", err)
	}
	if equipo.Codigo() != "C3-EQ01" {
		t.Fatalf("codigo no normalizado: %q", equipo.Codigo())
	}
	if equipo.Estado() != EquipoOperativo {
		t.Fatalf("estado inicial inesperado: %q", equipo.Estado())
	}
}

func TestDarDeAltaEquipoExigeTipo(t *testing.T) {
	_, err := DarDeAltaEquipo(identificador.Nuevo(), identificador.Nuevo(), identificador.Identificador{}, identificador.Nuevo(), "EQ", "", "", "", "", "", nil, nil, nil)
	if !errors.Is(err, ErrTipoDeEquipoObligatorio) {
		t.Fatalf("se esperaba ErrTipoDeEquipoObligatorio, se obtuvo %v", err)
	}
}

func TestDarDeAltaEquipoRechazaAnioFueraDeRango(t *testing.T) {
	anio := 1899
	_, err := DarDeAltaEquipo(identificador.Nuevo(), identificador.Nuevo(), identificador.Nuevo(), identificador.Nuevo(), "EQ", "", "", "", "", "", nil, &anio, nil)
	if !errors.Is(err, ErrValorFueraDeRango) {
		t.Fatalf("se esperaba ErrValorFueraDeRango, se obtuvo %v", err)
	}
}
