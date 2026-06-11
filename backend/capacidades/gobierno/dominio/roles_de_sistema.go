package dominio

type DefinicionDeRolDeSistema struct {
	Codigo      string
	Descripcion string
	Permisos    []string
}

const TodosLosPermisos = "*"

func RolesDeSistema() []DefinicionDeRolDeSistema {
	return []DefinicionDeRolDeSistema{
		{"ADMIN_EMPRESA", "Administra usuarios, roles y catalogos de su empresa", []string{TodosLosPermisos}},
		{"JEFE_TURNO", "Supervisa la operacion del turno y valida partes", []string{
			"produccion.ver", "produccion.capturar", "produccion.editar",
			"catalogos.ver", "planeacion.ver", "estandares.ver", "reportes.ver"}},
		{"CAPITAN_MINA", "Captura y valida partes de su mina", []string{
			"produccion.ver", "produccion.capturar", "catalogos.ver", "reportes.ver"}},
		{"OPERADOR", "Captura sus propios partes de operacion", []string{
			"produccion.ver", "produccion.capturar"}},
		{"PLANEACION", "Gestiona plan de bloques, metas y reportes", []string{
			"planeacion.ver", "planeacion.editar", "reconciliacion.ver", "reconciliacion.capturar",
			"costos.ver", "costos.editar", "estandares.ver", "estandares.editar",
			"catalogos.ver", "reportes.ver", "reportes.exportar"}},
		{"LECTURA", "Solo consulta tableros y reportes", []string{
			"catalogos.ver", "produccion.ver", "planeacion.ver", "reconciliacion.ver",
			"beneficio.ver", "estandares.ver", "costos.ver", "inversiones.ver", "reportes.ver"}},
	}
}
