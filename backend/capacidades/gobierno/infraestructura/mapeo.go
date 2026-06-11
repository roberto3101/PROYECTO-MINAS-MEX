package infraestructura

import "minas/compartido/identificador"

func textoOpcional(id *identificador.Identificador) any {
	if id == nil {
		return nil
	}
	return id.Texto()
}

func identificadorOpcional(texto *string) (*identificador.Identificador, error) {
	if texto == nil || *texto == "" {
		return nil, nil
	}
	convertido, err := identificador.Desde(*texto)
	if err != nil {
		return nil, err
	}
	return &convertido, nil
}

func identificadores(textos []string) ([]identificador.Identificador, error) {
	convertidos := make([]identificador.Identificador, 0, len(textos))
	for _, texto := range textos {
		convertido, err := identificador.Desde(texto)
		if err != nil {
			return nil, err
		}
		convertidos = append(convertidos, convertido)
	}
	return convertidos, nil
}

func textos(identificadores []identificador.Identificador) []string {
	convertidos := make([]string, 0, len(identificadores))
	for _, identificador := range identificadores {
		convertidos = append(convertidos, identificador.Texto())
	}
	return convertidos
}
