package aplicacion

import (
	"context"

	"minas/capacidades/seguridad/dominio"
	"minas/capacidades/seguridad/puertos"
	"minas/compartido/identificador"
	"minas/plataforma/escudo"
)

type ComandoReportarIncidente struct {
	IdentificadorEmpresa       string
	IdentificadorMina          string
	IdentificadorObra          string
	IdentificadorTipoIncidente string
	IdentificadorReportadoPor  string
	Fecha                      string
	Turno                      string
	Severidad                  string
	Descripcion                string
	AccionInmediata            string
}

type ReportarIncidente struct {
	unidad     puertos.UnidadDeTrabajo
	incidentes puertos.RepositorioDeIncidentes
}

func NuevoReportarIncidente(unidad puertos.UnidadDeTrabajo, incidentes puertos.RepositorioDeIncidentes) *ReportarIncidente {
	return &ReportarIncidente{unidad: unidad, incidentes: incidentes}
}

func (caso *ReportarIncidente) Ejecutar(ctx context.Context, comando ComandoReportarIncidente) (string, error) {
	empresa, err := identificador.Desde(comando.IdentificadorEmpresa)
	if err != nil {
		return "", err
	}
	mina, err := identificador.Desde(comando.IdentificadorMina)
	if err != nil {
		return "", err
	}
	tipoIncidente, err := identificador.Desde(comando.IdentificadorTipoIncidente)
	if err != nil {
		return "", err
	}
	obra, err := identificador.DesdeOpcional(comando.IdentificadorObra)
	if err != nil {
		return "", err
	}
	reportadoPor, err := identificador.DesdeOpcional(comando.IdentificadorReportadoPor)
	if err != nil {
		return "", err
	}
	if err := escudo.ValidarTextos(
		escudo.Campo("descripcion", &comando.Descripcion, 1000),
		escudo.Campo("accion inmediata", &comando.AccionInmediata, 1000),
	); err != nil {
		return "", err
	}
	incidente, err := dominio.ReportarIncidente(empresa, mina, tipoIncidente, obra, reportadoPor,
		comando.Fecha, comando.Turno, comando.Severidad, comando.Descripcion, comando.AccionInmediata)
	if err != nil {
		return "", err
	}
	if err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		return caso.incidentes.Guardar(ctx, incidente)
	}); err != nil {
		return "", err
	}
	return incidente.Identificador().Texto(), nil
}

type ConsultasDeSeguridad struct {
	unidad puertos.UnidadDeTrabajo
	lector puertos.LectorDeSeguridad
}

func NuevasConsultasDeSeguridad(unidad puertos.UnidadDeTrabajo, lector puertos.LectorDeSeguridad) *ConsultasDeSeguridad {
	return &ConsultasDeSeguridad{unidad: unidad, lector: lector}
}

func (caso *ConsultasDeSeguridad) ListarIncidentes(ctx context.Context, filtro puertos.FiltroDeIncidentes) ([]puertos.ResumenDeIncidente, string, error) {
	var incidentes []puertos.ResumenDeIncidente
	var siguiente string
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		listados, cursor, err := caso.lector.ListarIncidentes(ctx, filtro)
		incidentes, siguiente = listados, cursor
		return err
	})
	return incidentes, siguiente, err
}

func (caso *ConsultasDeSeguridad) ListarTiposDeIncidente(ctx context.Context) ([]puertos.OpcionDeSeguridad, error) {
	var tipos []puertos.OpcionDeSeguridad
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		listados, err := caso.lector.ListarTiposDeIncidente(ctx)
		tipos = listados
		return err
	})
	return tipos, err
}

func (caso *ConsultasDeSeguridad) ConteoPorSeveridad(ctx context.Context, filtro puertos.FiltroDeIncidentes) ([]puertos.ConteoPorSeveridad, error) {
	var conteo []puertos.ConteoPorSeveridad
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		resultado, err := caso.lector.ConteoPorSeveridad(ctx, filtro)
		conteo = resultado
		return err
	})
	return conteo, err
}
