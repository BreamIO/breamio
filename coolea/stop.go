package main

//Parse the subtree of stop
func parseStop(tokens []string) (string, handlerFunc, []string) {

	if len(tokens) > 1 {
		switch tokens[0] {
		case "et":
			return stopET(tokens[1:])
		}
	}

	return parseError(tokens[0])
}

func stopET(tokens []string) (string, handlerFunc, []string) {
	id, err := strconv.Atoi(tokens[0])
	if err != nil {
		fmt.Println(err)
		return parseError(tokens[0])
	}
	payload, err := json.Marshal(struct{}{})
	client.Send(aioli.ExtPkg{"tracker:shutdown", id, payload})
	return "Sent request to stop ET " + tokens[0], startParse, nil
}
