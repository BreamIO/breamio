package main

//Parse the subtree of create
func parseCreate(tokens []string) (string, handlerFunc, []string) {
	if len(tokens) > 1 {
		switch tokens[0] {
		case "region":
			return createRegion(tokens[1:])
		}
	}
	return parseError(tokens[0])
}

//Sends a message to start a region
func createRegion(tokens []string) (string, handlerFunc, []string) {

	//tokens[0] is name

	id, err := strconv.Atoi(tokens[1])
	if err != nil {
		fmt.Println(err)
		return parseError(tokens[0])
	}

	rd := regionStats.RegionDefinition{
		Type: "rect",
	}
	for i := 2; i < len(tokens); i += 2 {
		switch str[i] {
		case "type":
			rd.Type = str[i+1]
			continue
		case "x":
			rd.X = str2float(str[i+1])
			continue
		case "y":
			rd.Y = str2float(str[i+1])
			continue
		case "w":
			rd.Width = str2float(str[i+1])
			continue
		case "h":
			rd.Height = str2float(str[i+1])
			continue
		}
	}
	payload, err := json.Marshal(dregionStats.RegionDefinitionPackage{tokens[0], rd}) //tokens[0] is name
	if err != nil {
		fmt.Println(err)
		return parseError(tokens[0])
	}

	client.Send(aioli.ExtPkg{"regionStats:addRegion", id, payload})
	return "Sent request to start new ET", startParse, nil
}
