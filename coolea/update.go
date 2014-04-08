package main

func parseUpdate(tokens []string) (string, handlerFunc, []string) {

	switch tokens[0] {
	case "region":
		return updateRegion(tokens[1:])
	}

	return parseError(tokens[0])
}

func updateRegion(tokens []string) (string, handlerFunc, []string) {
	//TODO
	return nil, nil, nil
}
