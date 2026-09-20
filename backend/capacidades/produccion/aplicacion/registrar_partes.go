package aplicacion

import (
	"context"

	"minas/capacidades/produccion/dominio"
	"minas/capacidades/produccion/puertos"
	"minas/compartido/identificador"
	"minas/plataforma/escudo"
)

type ViajeCapturado struct {
	Desde                    string
	Hasta                    string
	IdentificadorTipoMineral string
	Toneladas                *float64
}

type ComandoRegistrarParteDeCarga struct {
	Modalidad               string
	IdentificadorEmpresa    string
	IdentificadorMina       string
	IdentificadorObra       string
	IdentificadorEquipo     string
	IdentificadorOperador   string
	IdentificadorSupervisor string
	Fecha                   string
	Turno                   string
	Observaciones           string
	HorometroInicial        float64
	HorometroFinal          float64
	Viajes                  []ViajeCapturado
}

type RegistrarParteDeCarga struct {
	unidad puertos.UnidadDeTrabajo
	partes puertos.RepositorioDeCarga
}

func NuevoRegistrarParteDeCarga(unidad puertos.UnidadDeTrabajo, partes puertos.RepositorioDeCarga) *RegistrarParteDeCarga {
	return &RegistrarParteDeCarga{unidad: unidad, partes: partes}
}

func (caso *RegistrarParteDeCarga) Ejecutar(ctx context.Context, comando ComandoRegistrarParteDeCarga) (string, error) {
	empresa, err := identificador.Desde(comando.IdentificadorEmpresa)
	if err != nil {
		return "", err
	}
	mina, err := identificador.Desde(comando.IdentificadorMina)
	if err != nil {
		return "", err
	}
	obra, err := identificador.Desde(comando.IdentificadorObra)
	if err != nil {
		return "", err
	}
	equipo, err := identificador.Desde(comando.IdentificadorEquipo)
	if err != nil {
		return "", err
	}
	operador, err := identificador.Desde(comando.IdentificadorOperador)
	if err != nil {
		return "", err
	}
	supervisor, err := identificador.DesdeOpcional(comando.IdentificadorSupervisor)
	if err != nil {
		return "", err
	}
	if err := escudo.ValidarTextos(escudo.Campo("observaciones", &comando.Observaciones, 500)); err != nil {
		return "", err
	}
	parte, err := dominio.RegistrarParteDeCarga(comando.Modalidad, empresa, mina, obra, equipo, operador,
		supervisor, comando.Fecha, comando.Turno, comando.Observaciones,
		comando.HorometroInicial, comando.HorometroFinal)
	if err != nil {
		return "", err
	}
	for _, viaje := range comando.Viajes {
		tipoMineral, err := identificador.Desde(viaje.IdentificadorTipoMineral)
		if err != nil {
			return "", err
		}
		if err := escudo.ValidarTextos(
			escudo.Campo("desde", &viaje.Desde, 80),
			escudo.Campo("hasta", &viaje.Hasta, 80),
		); err != nil {
			return "", err
		}
		if err := parte.AgregarViaje(viaje.Desde, viaje.Hasta, tipoMineral, viaje.Toneladas); err != nil {
			return "", err
		}
	}
	if err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		return caso.partes.Guardar(ctx, parte)
	}); err != nil {
		return "", err
	}
	return parte.Identificador().Texto(), nil
}

type AvanceCapturado struct {
	Actividad                string
	Lugar                    string
	NumeroDeSecciones        *int
	Longitud                 float64
	NumeroDeBarrenos         int
	IdentificadoresDeBarreno []string
}

type ComandoRegistrarParteDeBarrenacion struct {
	TipoDeBarrenacion        string
	IdentificadorEmpresa     string
	IdentificadorMina        string
	IdentificadorObra        string
	IdentificadorEquipo      string
	IdentificadorCapitanMina string
	IdentificadorOperador    string
	IdentificadorSupervisor  string
	IdentificadorAyudante    string
	IdentificadorActividad   string
	Fecha                    string
	Turno                    string
	Observaciones            string
	Horometros               dominio.Horometros
	Avances                  []AvanceCapturado
}

type RegistrarParteDeBarrenacion struct {
	unidad puertos.UnidadDeTrabajo
	partes puertos.RepositorioDeBarrenacion
}

func NuevoRegistrarParteDeBarrenacion(unidad puertos.UnidadDeTrabajo, partes puertos.RepositorioDeBarrenacion) *RegistrarParteDeBarrenacion {
	return &RegistrarParteDeBarrenacion{unidad: unidad, partes: partes}
}

func (caso *RegistrarParteDeBarrenacion) Ejecutar(ctx context.Context, comando ComandoRegistrarParteDeBarrenacion) (string, error) {
	empresa, err := identificador.Desde(comando.IdentificadorEmpresa)
	if err != nil {
		return "", err
	}
	mina, err := identificador.Desde(comando.IdentificadorMina)
	if err != nil {
		return "", err
	}
	obra, err := identificador.Desde(comando.IdentificadorObra)
	if err != nil {
		return "", err
	}
	equipo, err := identificador.Desde(comando.IdentificadorEquipo)
	if err != nil {
		return "", err
	}
	capitan, err := identificador.Desde(comando.IdentificadorCapitanMina)
	if err != nil {
		return "", err
	}
	operador, err := identificador.Desde(comando.IdentificadorOperador)
	if err != nil {
		return "", err
	}
	supervisor, err := identificador.DesdeOpcional(comando.IdentificadorSupervisor)
	if err != nil {
		return "", err
	}
	ayudante, err := identificador.DesdeOpcional(comando.IdentificadorAyudante)
	if err != nil {
		return "", err
	}
	actividad, err := identificador.DesdeOpcional(comando.IdentificadorActividad)
	if err != nil {
		return "", err
	}
	if err := escudo.ValidarTextos(escudo.Campo("observaciones", &comando.Observaciones, 500)); err != nil {
		return "", err
	}
	parte, err := dominio.RegistrarParteDeBarrenacion(comando.TipoDeBarrenacion, empresa, mina, obra, equipo,
		capitan, operador, supervisor, ayudante, actividad,
		comando.Fecha, comando.Turno, comando.Observaciones, comando.Horometros)
	if err != nil {
		return "", err
	}
	for _, avance := range comando.Avances {
		if err := escudo.ValidarTextos(escudo.Campo("lugar", &avance.Lugar, 80)); err != nil {
			return "", err
		}
		var ejecutados []identificador.Identificador
		for _, crudo := range avance.IdentificadoresDeBarreno {
			barreno, err := identificador.Desde(crudo)
			if err != nil {
				return "", err
			}
			ejecutados = append(ejecutados, barreno)
		}
		if err := parte.AgregarAvance(avance.Actividad, avance.Lugar, avance.NumeroDeSecciones,
			avance.Longitud, avance.NumeroDeBarrenos, ejecutados); err != nil {
			return "", err
		}
	}
	if err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		return caso.partes.Guardar(ctx, parte)
	}); err != nil {
		return "", err
	}
	return parte.Identificador().Texto(), nil
}

type ComandoRegistrarDemora struct {
	IdentificadorEmpresa    string
	IdentificadorMina       string
	IdentificadorEquipo     string
	IdentificadorTipoDemora string
	Fecha                   string
	Turno                   string
	Minutos                 int
	Observacion             string
}

type RegistrarDemora struct {
	unidad  puertos.UnidadDeTrabajo
	demoras puertos.RepositorioDeDemora
}

func NuevoRegistrarDemora(unidad puertos.UnidadDeTrabajo, demoras puertos.RepositorioDeDemora) *RegistrarDemora {
	return &RegistrarDemora{unidad: unidad, demoras: demoras}
}

func (caso *RegistrarDemora) Ejecutar(ctx context.Context, comando ComandoRegistrarDemora) (string, error) {
	empresa, err := identificador.Desde(comando.IdentificadorEmpresa)
	if err != nil {
		return "", err
	}
	mina, err := identificador.Desde(comando.IdentificadorMina)
	if err != nil {
		return "", err
	}
	equipo, err := identificador.Desde(comando.IdentificadorEquipo)
	if err != nil {
		return "", err
	}
	tipoDemora, err := identificador.Desde(comando.IdentificadorTipoDemora)
	if err != nil {
		return "", err
	}
	if err := escudo.ValidarTextos(escudo.Campo("observacion", &comando.Observacion, 300)); err != nil {
		return "", err
	}
	demora, err := dominio.RegistrarDemora(empresa, mina, equipo, tipoDemora,
		comando.Fecha, comando.Turno, comando.Minutos, comando.Observacion)
	if err != nil {
		return "", err
	}
	if err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		return caso.demoras.Guardar(ctx, demora)
	}); err != nil {
		return "", err
	}
	return demora.Identificador().Texto(), nil
}
