package infraestructura

import (
	"context"

	"minas/capacidades/catalogos/dominio"
	"minas/compartido/identificador"
	"minas/plataforma/persistencia"
)

type RepositorioObraPostgres struct{}

func NuevoRepositorioObra() RepositorioObraPostgres {
	return RepositorioObraPostgres{}
}

func (RepositorioObraPostgres) Guardar(ctx context.Context, obra dominio.Obra) error {
	consultas := persistencia.ConsultasDe(ctx)
	_, err := consultas.Exec(ctx,
		`INSERT INTO catalogos.obra (id, id_empresa, id_mina, id_tipo_obra, codigo, nombre, ubicacion, es_prioritaria)
		 VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), NULLIF($7, ''), $8)`,
		obra.Identificador().Texto(), obra.Empresa().Texto(), obra.Mina().Texto(),
		textoOpcionalDeIdentificador(obra.TipoDeObra()), obra.Codigo(), obra.Nombre(),
		obra.Ubicacion(), obra.EsPrioritaria())
	return traducirErrorDeEscritura(err)
}

func (RepositorioObraPostgres) CambiarEstado(ctx context.Context, id identificador.Identificador, estado string) error {
	consultas := persistencia.ConsultasDe(ctx)
	_, err := consultas.Exec(ctx,
		`UPDATE catalogos.obra SET estado = $2, actualizado_en = now() WHERE id = $1`,
		id.Texto(), estado)
	return err
}

func textoOpcionalDeIdentificador(valor *identificador.Identificador) any {
	if valor == nil {
		return nil
	}
	return valor.Texto()
}
