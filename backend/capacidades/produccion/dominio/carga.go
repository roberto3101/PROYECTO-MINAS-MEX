package dominio

import (
	"strings"

	"minas/compartido/identificador"
)

type Viaje struct {
	id            identificador.Identificador
	desde         string
	hasta         string
	idTipoMineral identificador.Identificador
	toneladas     *float64
}

func (viaje Viaje) Identificador() identificador.Identificador { return viaje.id }
func (viaje Viaje) Desde() string                              { return viaje.desde }
func (viaje Viaje) Hasta() string                              { return viaje.hasta }
func (viaje Viaje) TipoDeMineral() identificador.Identificador { return viaje.idTipoMineral }
func (viaje Viaje) Toneladas() *float64                        { return viaje.toneladas }

type ParteDeCarga struct {
	Cabecera
	modalidad        string
	idOperador       identificador.Identificador
	idSupervisor     *identificador.Identificador
	horometroInicial float64
	horometroFinal   float64
	viajes           []Viaje
}

func RegistrarParteDeCarga(modalidad string, idEmpresa, idMina, idObra, idEquipo, idOperador identificador.Identificador,
	idSupervisor *identificador.Identificador, fecha, turno, observaciones string,
	horometroInicial, horometroFinal float64) (ParteDeCarga, error) {
	if modalidad != ModalidadAcarreo && modalidad != ModalidadRezagado {
		return ParteDeCarga{}, ErrModalidadNoReconocida
	}
	cabecera, err := nuevaCabecera(idEmpresa, idMina, idObra, idEquipo, fecha, turno, observaciones)
	if err != nil {
		return ParteDeCarga{}, err
	}
	if err := validarHorometros(horometroInicial, horometroFinal); err != nil {
		return ParteDeCarga{}, err
	}
	return ParteDeCarga{
		Cabecera:         cabecera,
		modalidad:        modalidad,
		idOperador:       idOperador,
		idSupervisor:     idSupervisor,
		horometroInicial: horometroInicial,
		horometroFinal:   horometroFinal,
	}, nil
}

func (parte *ParteDeCarga) AgregarViaje(desde, hasta string, idTipoMineral identificador.Identificador, toneladas *float64) error {
	origen := strings.TrimSpace(desde)
	destino := strings.TrimSpace(hasta)
	if origen == "" || destino == "" {
		return ErrOrigenYDestinoObligatorios
	}
	if toneladas != nil && *toneladas <= 0 {
		return ErrToneladasInvalidas
	}
	parte.viajes = append(parte.viajes, Viaje{
		id:            identificador.Nuevo(),
		desde:         origen,
		hasta:         destino,
		idTipoMineral: idTipoMineral,
		toneladas:     toneladas,
	})
	return nil
}

func (parte ParteDeCarga) Modalidad() string                        { return parte.modalidad }
func (parte ParteDeCarga) Operador() identificador.Identificador    { return parte.idOperador }
func (parte ParteDeCarga) Supervisor() *identificador.Identificador { return parte.idSupervisor }
func (parte ParteDeCarga) HorometroInicial() float64                { return parte.horometroInicial }
func (parte ParteDeCarga) HorometroFinal() float64                  { return parte.horometroFinal }
func (parte ParteDeCarga) Viajes() []Viaje                          { return parte.viajes }

func (parte ParteDeCarga) HorasTrabajadas() float64 {
	return parte.horometroFinal - parte.horometroInicial
}

func (parte ParteDeCarga) ToneladasTotales() float64 {
	var total float64
	for _, viaje := range parte.viajes {
		if viaje.toneladas != nil {
			total += *viaje.toneladas
		}
	}
	return total
}
