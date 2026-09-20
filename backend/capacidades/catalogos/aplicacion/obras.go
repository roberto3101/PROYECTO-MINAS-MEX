package aplicacion

import (
	"context"

	"minas/capacidades/catalogos/dominio"
	"minas/capacidades/catalogos/puertos"
	"minas/compartido/identificador"
	"minas/plataforma/escudo"
)

type ComandoCrearObra struct {
	IdentificadorEmpresa  string
	IdentificadorMina     string
	IdentificadorTipoObra string
	Codigo                string
	Nombre                string
	Ubicacion             string
	EsPrioritaria         bool
}

type CrearObra struct {
	unidad puertos.UnidadDeTrabajo
	obras  puertos.RepositorioObra
}

func NuevoCrearObra(unidad puertos.UnidadDeTrabajo, obras puertos.RepositorioObra) *CrearObra {
	return &CrearObra{unidad: unidad, obras: obras}
}

func (caso *CrearObra) Ejecutar(ctx context.Context, comando ComandoCrearObra) (string, error) {
	empresa, err := identificador.Desde(comando.IdentificadorEmpresa)
	if err != nil {
		return "", err
	}
	mina, err := identificador.Desde(comando.IdentificadorMina)
	if err != nil {
		return "", dominio.ErrMinaObligatoria
	}
	tipoDeObra, err := identificador.DesdeOpcional(comando.IdentificadorTipoObra)
	if err != nil {
		return "", err
	}
	if err := escudo.ValidarTextos(
		escudo.Campo("codigo", &comando.Codigo, 40),
		escudo.Campo("nombre", &comando.Nombre, 120),
		escudo.Campo("ubicacion", &comando.Ubicacion, 80),
	); err != nil {
		return "", err
	}
	obra, err := dominio.CrearObra(empresa, mina, tipoDeObra, comando.Codigo, comando.Nombre,
		comando.Ubicacion, comando.EsPrioritaria)
	if err != nil {
		return "", err
	}
	if err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		return caso.obras.Guardar(ctx, obra)
	}); err != nil {
		return "", err
	}
	return obra.Identificador().Texto(), nil
}

type ListarObras struct {
	unidad puertos.UnidadDeTrabajo
	lector puertos.LectorDeCatalogos
}

func NuevoListarObras(unidad puertos.UnidadDeTrabajo, lector puertos.LectorDeCatalogos) *ListarObras {
	return &ListarObras{unidad: unidad, lector: lector}
}

func (caso *ListarObras) Ejecutar(ctx context.Context, filtro puertos.FiltroDeCatalogo) ([]puertos.ResumenObra, string, error) {
	var obras []puertos.ResumenObra
	var siguiente string
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		listadas, cursor, err := caso.lector.ListarObras(ctx, filtro)
		obras, siguiente = listadas, cursor
		return err
	})
	return obras, siguiente, err
}

type DetalleDeObra struct {
	unidad puertos.UnidadDeTrabajo
	lector puertos.LectorDeCatalogos
}

func NuevoDetalleDeObra(unidad puertos.UnidadDeTrabajo, lector puertos.LectorDeCatalogos) *DetalleDeObra {
	return &DetalleDeObra{unidad: unidad, lector: lector}
}

func (caso *DetalleDeObra) Ejecutar(ctx context.Context, identificadorObra string) (puertos.ResumenObra, bool, error) {
	var detalle puertos.ResumenObra
	var encontrada bool
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		resultado, existe, err := caso.lector.DetalleDeObra(ctx, identificadorObra)
		detalle, encontrada = resultado, existe
		return err
	})
	return detalle, encontrada, err
}

type ComandoCambiarEstadoDeObra struct {
	IdentificadorObra string
	Estado            string
}

type CambiarEstadoDeObra struct {
	unidad puertos.UnidadDeTrabajo
	obras  puertos.RepositorioObra
}

func NuevoCambiarEstadoDeObra(unidad puertos.UnidadDeTrabajo, obras puertos.RepositorioObra) *CambiarEstadoDeObra {
	return &CambiarEstadoDeObra{unidad: unidad, obras: obras}
}

func (caso *CambiarEstadoDeObra) Ejecutar(ctx context.Context, comando ComandoCambiarEstadoDeObra) error {
	id, err := identificador.Desde(comando.IdentificadorObra)
	if err != nil {
		return err
	}
	if !dominio.EsEstadoValido(comando.Estado, dominio.EstadosDeObraValidos) {
		return dominio.ErrEstadoNoReconocido
	}
	return caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		return caso.obras.CambiarEstado(ctx, id, comando.Estado)
	})
}
