package dominio

type DefinicionDeRolDeSistema struct {
	Codigo      string
	Descripcion string
	Permisos    []string
}

const TodosLosPermisos = "*"

const CodigoRolAdministrador = "ADMIN_EMPRESA"

func RolesDeSistema() []DefinicionDeRolDeSistema {
	return []DefinicionDeRolDeSistema{
		{CodigoRolAdministrador, "Administrador de la empresa: acceso total a todas las minas", []string{TodosLosPermisos}},
	}
}
