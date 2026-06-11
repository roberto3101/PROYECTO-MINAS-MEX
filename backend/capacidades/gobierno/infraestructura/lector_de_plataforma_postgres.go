package infraestructura

import (
	"context"

	"minas/capacidades/gobierno/puertos"
	"minas/compartido/paginacion"
	"minas/plataforma/persistencia"
)

type LectorDePlataformaPostgres struct{}

func NuevoLectorDePlataforma() LectorDePlataformaPostgres {
	return LectorDePlataformaPostgres{}
}

func (LectorDePlataformaPostgres) BuscarCredencialDeSuperadmin(ctx context.Context, nombreCorto string) (puertos.CredencialDeSuperadmin, bool, error) {
	consultas := persistencia.ConsultasDe(ctx)
	var credencial puertos.CredencialDeSuperadmin
	fila := consultas.QueryRow(ctx,
		`SELECT id, usuario, nombre, COALESCE(contrasena_hash, ''), (estado = 'ACTIVO')
		 FROM gobierno.superadmin WHERE usuario = $1 AND eliminado_en IS NULL`, nombreCorto)
	err := fila.Scan(&credencial.Identificador, &credencial.NombreCorto, &credencial.Nombre, &credencial.HashContrasena, &credencial.Activo)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return puertos.CredencialDeSuperadmin{}, false, nil
		}
		return puertos.CredencialDeSuperadmin{}, false, err
	}
	return credencial, true, nil
}

func (LectorDePlataformaPostgres) ListarEmpresas(ctx context.Context, filtro puertos.FiltroDeEmpresas) ([]puertos.ResumenEmpresaDePlataforma, string, error) {
	consultas := persistencia.ConsultasDe(ctx)
	orden, identificadorCursor, err := paginacion.DecodificarCursor(filtro.Cursor)
	if err != nil {
		return nil, "", err
	}
	filas, err := consultas.Query(ctx,
		`SELECT e.id, e.codigo, e.razon_social, e.estado,
		        COALESCE(e.logo_url,''), COALESCE(e.color_primario,''), e.moneda_defecto, e.zona_horaria,
		        to_char(e.creado_en, 'YYYY-MM-DD"T"HH24:MI:SSZ'),
		        (SELECT count(*) FROM gobierno.usuario u WHERE u.id_empresa = e.id AND u.eliminado_en IS NULL)
		 FROM gobierno.empresa e
		 WHERE e.eliminado_en IS NULL
		   AND ($1 = '' OR e.codigo ILIKE '%'||$1||'%' OR e.razon_social ILIKE '%'||$1||'%')
		   AND ($2 = '' OR $2 = 'TODAS' OR e.estado = $2)
		   AND ($3 = '' OR (e.codigo, e.id::text) > ($3, $4))
		 ORDER BY e.codigo, e.id LIMIT $5`,
		filtro.Busqueda, filtro.Estado, orden, identificadorCursor, filtro.Limite+1)
	if err != nil {
		return nil, "", err
	}
	defer filas.Close()
	var empresas []puertos.ResumenEmpresaDePlataforma
	for filas.Next() {
		var empresa puertos.ResumenEmpresaDePlataforma
		if err := filas.Scan(&empresa.Identificador, &empresa.Codigo, &empresa.RazonSocial, &empresa.Estado,
			&empresa.LogoUrl, &empresa.ColorPrimario, &empresa.Moneda, &empresa.ZonaHoraria,
			&empresa.CreadoEn, &empresa.TotalUsuarios); err != nil {
			return nil, "", err
		}
		empresas = append(empresas, empresa)
	}
	if err := filas.Err(); err != nil {
		return nil, "", err
	}
	siguiente := ""
	if len(empresas) > filtro.Limite {
		empresas = empresas[:filtro.Limite]
		ultima := empresas[len(empresas)-1]
		siguiente = paginacion.CodificarCursor(ultima.Codigo, ultima.Identificador)
	}
	return empresas, siguiente, nil
}

func (LectorDePlataformaPostgres) DetalleDeEmpresa(ctx context.Context, identificador string) (puertos.DetalleEmpresaDePlataforma, bool, error) {
	consultas := persistencia.ConsultasDe(ctx)
	var detalle puertos.DetalleEmpresaDePlataforma
	fila := consultas.QueryRow(ctx,
		`SELECT e.id, e.codigo, e.razon_social, e.estado,
		        COALESCE(e.logo_url,''), COALESCE(e.color_primario,''), e.moneda_defecto, e.zona_horaria,
		        to_char(e.creado_en, 'YYYY-MM-DD"T"HH24:MI:SSZ'),
		        COALESCE(e.identificacion_fiscal,''), COALESCE(e.correo_contacto,''), COALESCE(e.telefono,''),
		        (SELECT count(*) FROM gobierno.usuario u WHERE u.id_empresa = e.id AND u.eliminado_en IS NULL),
		        (SELECT count(*) FROM catalogos.mina m WHERE m.id_empresa = e.id AND m.eliminado_en IS NULL)
		 FROM gobierno.empresa e WHERE e.id = $1 AND e.eliminado_en IS NULL`, identificador)
	err := fila.Scan(&detalle.Identificador, &detalle.Codigo, &detalle.RazonSocial, &detalle.Estado,
		&detalle.LogoUrl, &detalle.ColorPrimario, &detalle.Moneda, &detalle.ZonaHoraria, &detalle.CreadoEn,
		&detalle.IdentificacionFiscal, &detalle.CorreoContacto, &detalle.Telefono,
		&detalle.TotalUsuarios, &detalle.TotalMinas)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return puertos.DetalleEmpresaDePlataforma{}, false, nil
		}
		return puertos.DetalleEmpresaDePlataforma{}, false, err
	}
	return detalle, true, nil
}

func (LectorDePlataformaPostgres) ExisteCodigoDeEmpresa(ctx context.Context, codigo string) (bool, error) {
	consultas := persistencia.ConsultasDe(ctx)
	var existe bool
	err := consultas.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM gobierno.empresa WHERE codigo = $1 AND eliminado_en IS NULL)`, codigo).Scan(&existe)
	return existe, err
}
