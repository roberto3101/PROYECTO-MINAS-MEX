package reloj

import "time"

type Reloj interface {
	Ahora() time.Time
}

type relojDelSistema struct{}

func DelSistema() Reloj {
	return relojDelSistema{}
}

func (relojDelSistema) Ahora() time.Time {
	return time.Now().UTC()
}
