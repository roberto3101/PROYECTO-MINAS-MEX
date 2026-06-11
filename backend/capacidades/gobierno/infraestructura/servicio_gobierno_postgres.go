package infraestructura

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"minas/capacidades/gobierno/contrato"
	"minas/capacidades/gobierno/puertos"
	"minas/plataforma/persistencia"
)

type ServicioGobiernoPostgres struct {
	unidad puertos.UnidadDeTrabajo
}

func NuevoServicioGobierno(unidad puertos.UnidadDeTrabajo) ServicioGobiernoPostgres {
	return ServicioGobiernoPostgres{unidad: unidad}
}

func (servicio ServicioGobiernoPostgres) PermisosVigentesDe(ctx context.Context, identificadorUsuario string) ([]contrato.PermisoVigente, error) {
	var vigentes []contrato.PermisoVigente
	err := servicio.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		consultas := persistencia.ConsultasDe(ctx)
		filas, err := consultas.Query(ctx,
			"SELECT permiso, modulo, COALESCE(alcance_mina::text, '') FROM gobierno.v_permisos_usuario WHERE id_usuario = $1",
			identificadorUsuario)
		if err != nil {
			return err
		}
		defer filas.Close()
		for filas.Next() {
			var permiso contrato.PermisoVigente
			if err := filas.Scan(&permiso.Codigo, &permiso.Modulo, &permiso.AlcanceMina); err != nil {
				return err
			}
			vigentes = append(vigentes, permiso)
		}
		return filas.Err()
	})
	if err != nil {
		return nil, err
	}
	return vigentes, nil
}

func (servicio ServicioGobiernoPostgres) EmpresaActual(ctx context.Context) (contrato.ResumenEmpresa, bool, error) {
	var resumen contrato.ResumenEmpresa
	encontrada := false
	err := servicio.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		consultas := persistencia.ConsultasDe(ctx)
		fila := consultas.QueryRow(ctx,
			`SELECT id, codigo, razon_social, (estado = 'ACTIVA'), zona_horaria, moneda_defecto,
			        COALESCE(logo_url, ''), COALESCE(color_primario, ''),
			        COALESCE(identificacion_fiscal, ''), COALESCE(correo_contacto, ''), COALESCE(telefono, '')
			 FROM gobierno.empresa LIMIT 1`)
		err := fila.Scan(&resumen.Identificador, &resumen.Codigo, &resumen.RazonSocial, &resumen.Activa,
			&resumen.ZonaHoraria, &resumen.Moneda, &resumen.LogoUrl, &resumen.ColorPrimario,
			&resumen.IdentificacionFiscal, &resumen.CorreoContacto, &resumen.Telefono)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		encontrada = true
		return nil
	})
	if err != nil {
		return contrato.ResumenEmpresa{}, false, err
	}
	return resumen, encontrada, nil
}
