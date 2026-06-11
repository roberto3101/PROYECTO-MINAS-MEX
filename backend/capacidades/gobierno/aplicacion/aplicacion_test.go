package aplicacion

import (
	"context"
	"errors"
	"testing"

	"minas/capacidades/gobierno/dominio"
	"minas/compartido/identificador"
)

type unidadDirecta struct{}

func (unidadDirecta) EnTransaccion(ctx context.Context, operacion func(context.Context) error) error {
	return operacion(ctx)
}

type cifradorSimulado struct{}

func (cifradorSimulado) Cifrar(textoPlano string) (string, error) {
	return "cifrado:" + textoPlano, nil
}

func (cifradorSimulado) Verificar(textoPlano, cifrado string) bool {
	return cifrado == "cifrado:"+textoPlano
}

type repositorioUsuarioMemoria struct {
	porIdentificador map[string]dominio.Usuario
}

func nuevoRepositorioUsuarioMemoria() *repositorioUsuarioMemoria {
	return &repositorioUsuarioMemoria{porIdentificador: map[string]dominio.Usuario{}}
}

func (repositorio *repositorioUsuarioMemoria) Guardar(_ context.Context, usuario dominio.Usuario) error {
	repositorio.porIdentificador[usuario.Identificador().Texto()] = usuario
	return nil
}

func (repositorio *repositorioUsuarioMemoria) BuscarPorIdentificador(_ context.Context, id identificador.Identificador) (dominio.Usuario, bool, error) {
	usuario, encontrado := repositorio.porIdentificador[id.Texto()]
	return usuario, encontrado, nil
}

func (repositorio *repositorioUsuarioMemoria) BuscarPorNombreCorto(_ context.Context, nombreCorto string) (dominio.Usuario, bool, error) {
	for _, usuario := range repositorio.porIdentificador {
		if usuario.NombreCorto() == nombreCorto {
			return usuario, true, nil
		}
	}
	return dominio.Usuario{}, false, nil
}

type repositorioAsignacionMemoria struct {
	porIdentificador map[string]dominio.AsignacionRol
}

func nuevoRepositorioAsignacionMemoria() *repositorioAsignacionMemoria {
	return &repositorioAsignacionMemoria{porIdentificador: map[string]dominio.AsignacionRol{}}
}

func (repositorio *repositorioAsignacionMemoria) Guardar(_ context.Context, asignacion dominio.AsignacionRol) error {
	repositorio.porIdentificador[asignacion.Identificador().Texto()] = asignacion
	return nil
}

func (repositorio *repositorioAsignacionMemoria) BuscarPorIdentificador(_ context.Context, id identificador.Identificador) (dominio.AsignacionRol, bool, error) {
	asignacion, encontrada := repositorio.porIdentificador[id.Texto()]
	return asignacion, encontrada, nil
}

func (repositorio *repositorioAsignacionMemoria) BuscarVigente(_ context.Context, _, _ identificador.Identificador, _ *identificador.Identificador) (dominio.AsignacionRol, bool, error) {
	return dominio.AsignacionRol{}, false, nil
}

func TestRegistrarUsuarioGuardaYDevuelveIdentificador(t *testing.T) {
	usuarios := nuevoRepositorioUsuarioMemoria()
	caso := NuevoRegistrarUsuario(unidadDirecta{}, usuarios, cifradorSimulado{})
	identificadorUsuario, err := caso.Ejecutar(context.Background(), ComandoRegistrarUsuario{
		IdentificadorEmpresa: identificador.Nuevo().Texto(),
		NombreCorto:          "op.tres",
		Nombre:               "Operador Tres",
		Contrasena:           "Pruebas#2026X",
	})
	if err != nil {
		t.Fatalf("registro fallido: %v", err)
	}
	if identificadorUsuario == "" {
		t.Fatal("se esperaba un identificador no vacio")
	}
	if len(usuarios.porIdentificador) != 1 {
		t.Fatalf("se esperaba 1 usuario almacenado, hay %d", len(usuarios.porIdentificador))
	}
}

func TestRegistrarUsuarioRechazaDuplicadoActivo(t *testing.T) {
	usuarios := nuevoRepositorioUsuarioMemoria()
	caso := NuevoRegistrarUsuario(unidadDirecta{}, usuarios, cifradorSimulado{})
	empresa := identificador.Nuevo().Texto()
	if _, err := caso.Ejecutar(context.Background(), ComandoRegistrarUsuario{IdentificadorEmpresa: empresa, NombreCorto: "op.tres", Nombre: "Operador Tres", Contrasena: "Pruebas#2026X"}); err != nil {
		t.Fatalf("primer registro fallido: %v", err)
	}
	_, err := caso.Ejecutar(context.Background(), ComandoRegistrarUsuario{IdentificadorEmpresa: empresa, NombreCorto: "op.tres", Nombre: "Otro Operador", Contrasena: "Pruebas#2026X"})
	if !errors.Is(err, ErrUsuarioDuplicado) {
		t.Fatalf("se esperaba ErrUsuarioDuplicado, se obtuvo %v", err)
	}
}

func TestAsignarRolGuardaLaAsignacion(t *testing.T) {
	asignaciones := nuevoRepositorioAsignacionMemoria()
	caso := NuevoAsignarRol(unidadDirecta{}, asignaciones)
	identificadorAsignacion, err := caso.Ejecutar(context.Background(), ComandoAsignarRol{
		IdentificadorEmpresa: identificador.Nuevo().Texto(),
		IdentificadorUsuario: identificador.Nuevo().Texto(),
		IdentificadorRol:     identificador.Nuevo().Texto(),
	})
	if err != nil {
		t.Fatalf("asignacion fallida: %v", err)
	}
	if identificadorAsignacion == "" {
		t.Fatal("se esperaba un identificador de asignacion no vacio")
	}
	if len(asignaciones.porIdentificador) != 1 {
		t.Fatalf("se esperaba 1 asignacion almacenada, hay %d", len(asignaciones.porIdentificador))
	}
}
