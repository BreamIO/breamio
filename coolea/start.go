package main

//Parse the subtree of start
func parseStart(tokens []string) (string, handlerFunc, []string) {
	if len(tokens) > 1 {
		switch tokens[0] {
		case "et":
			return startET(tokens[1:])
		}
	}
	return parseError(tokens[0])
}

//Sends a start et message
func startET(tokens []string) (string, handlerFunc, []string) {
	id, err := strconv.Atoi(tokens[0])
	if err != nil {
		fmt.Println(err)
		return parseError(tokens[0])
	}
	payload, err := json.Marshal(beenleigh.Spec{id, tokens[1]})
	if err != nil {
		fmt.Println(err)
		return parseError(tokens[1])
	}
	client.Send(aioli.ExtPkg{"new:tracker", 256, payload})
	return "Sent request to start new ET", startParse, nil
}
