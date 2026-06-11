//go:build integracion

package integracion

import (
	"context"
	"os"
	"testing"

	"minas/capacidades/gobierno/aplicacion"
	"minas/capacidades/gobierno/infraestructura"
	"minas/compartido/identificador"
	"minas/compartido/reloj"
	"minas/plataforma/contexto"
	"minas/plataforma/escudo"
	"minas/plataforma/persistencia"
	"minas/plataforma/seguridad"
)

const identificadorEmpresaMin2 = "99999999-9999-9999-9999-999999999999"

func cadenaDeConexion() string {
	return os.Getenv("CADENA_POSTGRES")
}

func TestFlujoGobiernoContraPostgres(t *testing.T) {
	if cadenaDeConexion() == "" {
		t.Skip("CADENA_POSTGRES no definida; se omite la prueba de integracion")
	}
	ctx := context.Background()
	pool, err := persistencia.NuevoPool(ctx, cadenaDeConexion())
	if err != nil {
		t.Fatalf("conexion fallida: %v", err)
	}
	defer pool.Close()

	cifrador := seguridad.NuevoCifradorBcrypt()
	hash, err := cifrador.Cifrar("clave-admin")
	if err != nil {
		t.Fatalf("cifrado fallido: %v", err)
	}
	var hashOriginal string
	if err := pool.QueryRow(ctx, "SELECT COALESCE(contrasena_hash, '') FROM gobierno.usuario WHERE usuario = 'admin.mina'").Scan(&hashOriginal); err != nil {
		t.Fatalf("lectura de credencial original fallida: %v", err)
	}
	if _, err := pool.Exec(ctx, "UPDATE gobierno.usuario SET contrasena_hash = $1 WHERE usuario = 'admin.mina'", hash); err != nil {
		t.Fatalf("preparacion de credencial fallida: %v", err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, "UPDATE gobierno.usuario SET contrasena_hash = NULLIF($1, '') WHERE usuario = 'admin.mina'", hashOriginal)
	}()
	if _, err := pool.Exec(ctx, "DELETE FROM gobierno.usuario WHERE usuario = 'op.integracion'"); err != nil {
		t.Fatalf("limpieza previa fallida: %v", err)
	}

	unidad := persistencia.NuevaUnidadDeTrabajo(pool)
	unidadDePlataforma := persistencia.NuevaUnidadDePlataforma(pool)
	repositorioUsuario := infraestructura.NuevoRepositorioUsuario()
	lectorDeAcceso := infraestructura.NuevoLectorDeAcceso()

	iniciarSesion := aplicacion.NuevoIniciarSesion(unidadDePlataforma, lectorDeAcceso, cifrador, escudo.NuevoLimitadorDeIntentos(), reloj.DelSistema())
	sesion, err := iniciarSesion.Ejecutar(ctx, aplicacion.ComandoIniciarSesion{
		CodigoEmpresa:     "MIN",
		NombreCorto:       "admin.mina",
		Contrasena:        "clave-admin",
		ClaveDeLimitacion: "integracion|local",
	})
	if err != nil {
		t.Fatalf("inicio de sesion fallido: %v", err)
	}
	if len(sesion.Permisos) < 29 {
		t.Fatalf("admin.mina deberia tener al menos 29 permisos vigentes, tiene %d", len(sesion.Permisos))
	}

	empresa, err := identificador.Desde(sesion.IdentificadorEmpresa)
	if err != nil {
		t.Fatalf("empresa de sesion invalida: %v", err)
	}
	actor, err := identificador.Desde(sesion.IdentificadorUsuario)
	if err != nil {
		t.Fatalf("actor de sesion invalido: %v", err)
	}

	ctxEmpresa := contexto.ConTenant(ctx, contexto.Tenant{Empresa: empresa, Actor: actor, Rol: contexto.RolAplicacion})
	registrarUsuario := aplicacion.NuevoRegistrarUsuario(unidad, repositorioUsuario, cifrador)
	identificadorNuevo, err := registrarUsuario.Ejecutar(ctxEmpresa, aplicacion.ComandoRegistrarUsuario{
		IdentificadorEmpresa: empresa.Texto(),
		NombreCorto:          "op.integracion",
		Nombre:               "Operador De Integracion",
		Contrasena:           "Integracion#2026",
	})
	if err != nil {
		t.Fatalf("registro de usuario fallido: %v", err)
	}

	usuarioCreado, err := identificador.Desde(identificadorNuevo)
	if err != nil {
		t.Fatalf("identificador de usuario invalido: %v", err)
	}

	visibleEnSuEmpresa := false
	if err := unidad.EnTransaccion(ctxEmpresa, func(ctx context.Context) error {
		_, encontrado, err := repositorioUsuario.BuscarPorIdentificador(ctx, usuarioCreado)
		visibleEnSuEmpresa = encontrado
		return err
	}); err != nil {
		t.Fatalf("lectura en la empresa propia fallida: %v", err)
	}
	if !visibleEnSuEmpresa {
		t.Fatal("el usuario recien creado deberia ser visible en su empresa")
	}

	empresaAjena, err := identificador.Desde(identificadorEmpresaMin2)
	if err != nil {
		t.Fatalf("empresa ajena invalida: %v", err)
	}
	ctxEmpresaAjena := contexto.ConTenant(ctx, contexto.Tenant{Empresa: empresaAjena, Actor: empresaAjena, Rol: contexto.RolAplicacion})
	visibleEnOtraEmpresa := true
	if err := unidad.EnTransaccion(ctxEmpresaAjena, func(ctx context.Context) error {
		_, encontrado, err := repositorioUsuario.BuscarPorIdentificador(ctx, usuarioCreado)
		visibleEnOtraEmpresa = encontrado
		return err
	}); err != nil {
		t.Fatalf("lectura en empresa ajena fallida: %v", err)
	}
	if visibleEnOtraEmpresa {
		t.Fatal("el aislamiento RLS fallo: otra empresa pudo ver el usuario")
	}
}
