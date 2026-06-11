package aplicacion

import (
	"context"

	"minas/capacidades/gobierno/dominio"
	"minas/capacidades/gobierno/puertos"
	"minas/compartido/identificador"
)

type ComandoCambiarEstadoDeEmpresa struct {
	IdentificadorEmpresa string
	Estado               string
}

type CambiarEstadoDeEmpresa struct {
	unidad   puertos.UnidadDeTrabajoDePlataforma
	empresas puertos.RepositorioEmpresa
}

func NuevoCambiarEstadoDeEmpresa(unidad puertos.UnidadDeTrabajoDePlataforma, empresas puertos.RepositorioEmpresa) *CambiarEstadoDeEmpresa {
	return &CambiarEstadoDeEmpresa{unidad: unidad, empresas: empresas}
}

func (caso *CambiarEstadoDeEmpresa) Ejecutar(ctx context.Context, comando ComandoCambiarEstadoDeEmpresa) error {
	idEmpresa, err := identificador.Desde(comando.IdentificadorEmpresa)
	if err != nil {
		return err
	}
	if comando.Estado != string(dominio.EmpresaActiva) && comando.Estado != string(dominio.EmpresaInactiva) {
		return dominio.ErrEstadoNoReconocido
	}
	return caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		_, encontrada, err := caso.empresas.BuscarPorIdentificador(ctx, idEmpresa)
		if err != nil {
			return err
		}
		if !encontrada {
			return ErrEmpresaNoEncontrada
		}
		return caso.empresas.CambiarEstado(ctx, idEmpresa, comando.Estado)
	})
}
