package contrato

import "context"

type MinaPublicada struct {
	Identificador string
	Nombre        string
}

type Catalogos interface {
	MinasActivas(ctx context.Context) ([]MinaPublicada, error)
	SembrarCatalogosBasicos(ctx context.Context, identificadorEmpresa string) error
}
