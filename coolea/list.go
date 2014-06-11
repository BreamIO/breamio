package main

//Parse the subtree of list
func parseList(tokens []string) (string, handlerFunc, []string) {
	return tokens[0], startParse, tokens[1:]
}
