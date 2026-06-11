package aplicacion

import (
	"context"
	"errors"

	contratoCatalogos "minas/capacidades/catalogos/contrato"
	"minas/capacidades/gobierno/dominio"
	"minas/capacidades/gobierno/puertos"
	"minas/compartido/identificador"
	"minas/plataforma/contexto"
	"minas/plataforma/escudo"
)

var ErrCodigoDeEmpresaEnUso = errors.New("ya existe una empresa activa con ese codigo")

type ComandoAprovisionarEmpresa struct {
	Codigo               string
	RazonSocial          string
	IdentificacionFiscal string
	CorreoContacto       string
	Telefono             string
	ZonaHoraria          string
	Moneda               string
	ColorPrimario        string
	AdminUsuario         string
	AdminNombre          string
	AdminCorreo          string
	AdminContrasena      string
}

type EmpresaAprovisionada struct {
	IdentificadorEmpresa string
	IdentificadorAdmin   string
}

type AprovisionarEmpresa struct {
	unidadDePlataforma puertos.UnidadDeTrabajoDePlataforma
	unidadDeTenant     puertos.UnidadDeTrabajo
	lector             puertos.LectorDePlataforma
	empresas           puertos.RepositorioEmpresa
	usuarios           puertos.RepositorioUsuario
	asignaciones       puertos.RepositorioAsignacionRol
	roles              puertos.RepositorioRol
	accesos            puertos.AprovisionadorDeAccesos
	cifrador           puertos.CifradorDeContrasena
	catalogos          contratoCatalogos.Catalogos
}

func NuevoAprovisionarEmpresa(
	unidadDePlataforma puertos.UnidadDeTrabajoDePlataforma,
	unidadDeTenant puertos.UnidadDeTrabajo,
	lector puertos.LectorDePlataforma,
	empresas puertos.RepositorioEmpresa,
	usuarios puertos.RepositorioUsuario,
	asignaciones puertos.RepositorioAsignacionRol,
	roles puertos.RepositorioRol,
	accesos puertos.AprovisionadorDeAccesos,
	cifrador puertos.CifradorDeContrasena,
	catalogos contratoCatalogos.Catalogos,
) *AprovisionarEmpresa {
	return &AprovisionarEmpresa{
		unidadDePlataforma: unidadDePlataforma, unidadDeTenant: unidadDeTenant,
		lector: lector, empresas: empresas, usuarios: usuarios, asignaciones: asignaciones,
		roles: roles, accesos: accesos, cifrador: cifrador, catalogos: catalogos,
	}
}

func (caso *AprovisionarEmpresa) Ejecutar(ctx context.Context, comando ComandoAprovisionarEmpresa) (EmpresaAprovisionada, error) {
	if err := escudo.ValidarCodigoDeEmpresa(comando.Codigo); err != nil {
		return EmpresaAprovisionada{}, err
	}
	if err := escudo.ValidarNombreDeUsuario(comando.AdminUsuario); err != nil {
		return EmpresaAprovisionada{}, err
	}
	if err := escudo.ValidarContrasena(comando.AdminContrasena); err != nil {
		return EmpresaAprovisionada{}, err
	}
	if err := escudo.ValidarCorreo(comando.AdminCorreo); err != nil {
		return EmpresaAprovisionada{}, err
	}
	if err := escudo.ValidarCorreo(comando.CorreoContacto); err != nil {
		return EmpresaAprovisionada{}, err
	}
	branding, err := dominio.ConfigurarBranding("", comando.ColorPrimario, comando.ZonaHoraria, comando.Moneda)
	if err != nil {
		return EmpresaAprovisionada{}, err
	}
	empresa, err := dominio.CrearEmpresa(comando.Codigo, comando.RazonSocial, dominio.PerfilDeContacto{
		IdentificacionFiscal: comando.IdentificacionFiscal,
		CorreoContacto:       comando.CorreoContacto,
		Telefono:             comando.Telefono,
	}, branding)
	if err != nil {
		return EmpresaAprovisionada{}, err
	}
	admin, err := dominio.RegistrarUsuario(empresa.Identificador(), comando.AdminUsuario, comando.AdminNombre)
	if err != nil {
		return EmpresaAprovisionada{}, err
	}
	if comando.AdminCorreo != "" {
		correo, err := dominio.NuevoCorreo(comando.AdminCorreo)
		if err != nil {
			return EmpresaAprovisionada{}, err
		}
		admin.DefinirCorreo(correo)
	}
	cifrada, err := caso.cifrador.Cifrar(comando.AdminContrasena)
	if err != nil {
		return EmpresaAprovisionada{}, err
	}
	admin.DefinirContrasenaCifrada(cifrada)

	err = caso.unidadDePlataforma.EnTransaccion(ctx, func(ctx context.Context) error {
		enUso, err := caso.lector.ExisteCodigoDeEmpresa(ctx, empresa.Codigo())
		if err != nil {
			return err
		}
		if enUso {
			return ErrCodigoDeEmpresaEnUso
		}
		if err := caso.empresas.Crear(ctx, empresa); err != nil {
			return err
		}
		if err := caso.usuarios.Guardar(ctx, admin); err != nil {
			return err
		}
		if err := caso.accesos.SembrarRolesDeSistema(ctx, empresa.Identificador(), dominio.RolesDeSistema()); err != nil {
			return err
		}
		rolAdmin, encontrado, err := caso.roles.BuscarPorCodigo(ctx, "ADMIN_EMPRESA")
		if err != nil {
			return err
		}
		if !encontrado {
			return ErrRolNoEncontrado
		}
		asignacion := dominio.AsignarRol(empresa.Identificador(), admin.Identificador(), rolAdmin.Identificador(), nil)
		return caso.asignaciones.Guardar(ctx, asignacion)
	})
	if err != nil {
		return EmpresaAprovisionada{}, err
	}

	tenantNuevo := contexto.Tenant{Empresa: empresa.Identificador(), Actor: admin.Identificador(), Rol: contexto.RolAplicacion}
	err = caso.unidadDeTenant.EnTransaccion(contexto.ConTenant(ctx, tenantNuevo), func(ctx context.Context) error {
		return caso.catalogos.SembrarCatalogosBasicos(ctx, empresa.Identificador().Texto())
	})
	if err != nil {
		return EmpresaAprovisionada{}, err
	}
	return EmpresaAprovisionada{
		IdentificadorEmpresa: empresa.Identificador().Texto(),
		IdentificadorAdmin:   admin.Identificador().Texto(),
	}, nil
}

var _ = identificador.Nuevo
