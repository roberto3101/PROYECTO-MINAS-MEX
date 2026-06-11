package web

import (
	"net/http"
	"strings"

	"minas/compartido/identificador"
	"minas/compartido/reloj"
	"minas/plataforma/contexto"
	"minas/plataforma/identidad"
)

type Autenticador struct {
	emisor identidad.EmisorDeToken
	reloj  reloj.Reloj
}

func NuevoAutenticador(emisor identidad.EmisorDeToken, reloj reloj.Reloj) Autenticador {
	return Autenticador{emisor: emisor, reloj: reloj}
}

func (autenticador Autenticador) Requerir(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(escritor http.ResponseWriter, peticion *http.Request) {
		sesion, err := autenticador.emisor.Verificar(tokenPortador(peticion), autenticador.reloj.Ahora())
		if err != nil {
			ResponderError(escritor, http.StatusUnauthorized, "no autorizado")
			return
		}
		empresa, err := identificador.Desde(sesion.IdentificadorEmpresa)
		if err != nil {
			ResponderError(escritor, http.StatusUnauthorized, "no autorizado")
			return
		}
		actor, err := identificador.Desde(sesion.IdentificadorUsuario)
		if err != nil {
			ResponderError(escritor, http.StatusUnauthorized, "no autorizado")
			return
		}
		tenant := contexto.Tenant{Empresa: empresa, Actor: actor, Rol: contexto.RolAplicacion}
		ctx := conSesion(contexto.ConTenant(peticion.Context(), tenant), sesion)
		siguiente.ServeHTTP(escritor, peticion.WithContext(ctx))
	})
}

func (autenticador Autenticador) Exigir(permiso string, manejador http.HandlerFunc) http.Handler {
	return autenticador.Requerir(http.HandlerFunc(func(escritor http.ResponseWriter, peticion *http.Request) {
		sesion, _ := SesionDe(peticion.Context())
		if !concedido(sesion.Permisos, permiso) {
			ResponderError(escritor, http.StatusForbidden, "permiso insuficiente: "+permiso)
			return
		}
		manejador(escritor, peticion)
	}))
}

func tokenPortador(peticion *http.Request) string {
	encabezado := peticion.Header.Get("Authorization")
	return strings.TrimPrefix(encabezado, "Bearer ")
}

func concedido(permisos []string, requerido string) bool {
	for _, permiso := range permisos {
		if permiso == requerido {
			return true
		}
	}
	return false
}
