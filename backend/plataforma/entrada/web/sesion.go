package web

import (
	"context"

	"minas/plataforma/identidad"
)

type claveSesion int

const claveSesionActual claveSesion = iota

func conSesion(ctx context.Context, sesion identidad.Sesion) context.Context {
	return context.WithValue(ctx, claveSesionActual, sesion)
}

func SesionDe(ctx context.Context) (identidad.Sesion, bool) {
	sesion, presente := ctx.Value(claveSesionActual).(identidad.Sesion)
	return sesion, presente
}
