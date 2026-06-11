package escudo

import (
	"sync"
	"time"
)

const (
	fallosAntesDelBloqueo = 5
	bloqueoInicial        = time.Minute
	bloqueoMaximo         = 15 * time.Minute
	ventanaDeFallos       = 15 * time.Minute
)

type registroDeIntentos struct {
	fallos         int
	ultimoFallo    time.Time
	bloqueadoHasta time.Time
	bloqueoActual  time.Duration
}

type LimitadorDeIntentos struct {
	candado   sync.Mutex
	registros map[string]*registroDeIntentos
}

func NuevoLimitadorDeIntentos() *LimitadorDeIntentos {
	return &LimitadorDeIntentos{registros: make(map[string]*registroDeIntentos)}
}

func (limitador *LimitadorDeIntentos) Permitir(clave string, ahora time.Time) (int, bool) {
	limitador.candado.Lock()
	defer limitador.candado.Unlock()
	registro, existe := limitador.registros[clave]
	if !existe {
		return 0, true
	}
	if ahora.Before(registro.bloqueadoHasta) {
		return int(registro.bloqueadoHasta.Sub(ahora).Seconds()) + 1, false
	}
	if ahora.Sub(registro.ultimoFallo) > ventanaDeFallos {
		delete(limitador.registros, clave)
	}
	return 0, true
}

func (limitador *LimitadorDeIntentos) RegistrarFallo(clave string, ahora time.Time) {
	limitador.candado.Lock()
	defer limitador.candado.Unlock()
	registro, existe := limitador.registros[clave]
	if !existe || ahora.Sub(registro.ultimoFallo) > ventanaDeFallos {
		registro = &registroDeIntentos{}
		limitador.registros[clave] = registro
	}
	registro.fallos++
	registro.ultimoFallo = ahora
	if registro.fallos >= fallosAntesDelBloqueo {
		if registro.bloqueoActual == 0 {
			registro.bloqueoActual = bloqueoInicial
		} else if registro.bloqueoActual < bloqueoMaximo {
			registro.bloqueoActual *= 2
		}
		registro.bloqueadoHasta = ahora.Add(registro.bloqueoActual)
		registro.fallos = 0
	}
}

func (limitador *LimitadorDeIntentos) RegistrarExito(clave string) {
	limitador.candado.Lock()
	defer limitador.candado.Unlock()
	delete(limitador.registros, clave)
}
