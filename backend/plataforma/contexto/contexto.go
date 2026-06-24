package contexto

import (
	"context"

	"minas/compartido/identificador"
)

type RolDeBaseDeDatos string

const (
	RolAplicacion RolDeBaseDeDatos = "aplicacion"
	RolPlataforma RolDeBaseDeDatos = "plataforma"
)

type Tenant struct {
	Empresa              identificador.Identificador
	Actor                identificador.Identificador
	Rol                  RolDeBaseDeDatos
	AlcanceGlobalDeMinas bool
	MinasPermitidas      []string
}

type clave int

const claveTenant clave = iota

func ConTenant(ctx context.Context, tenant Tenant) context.Context {
	return context.WithValue(ctx, claveTenant, tenant)
}

func ConEmpresaImpersonada(ctx context.Context, empresa identificador.Identificador) context.Context {
	return ConTenant(ctx, Tenant{
		Empresa:              empresa,
		Actor:                empresa,
		Rol:                  RolAplicacion,
		AlcanceGlobalDeMinas: true,
	})
}

func TenantDe(ctx context.Context) (Tenant, bool) {
	tenant, presente := ctx.Value(claveTenant).(Tenant)
	return tenant, presente
}
