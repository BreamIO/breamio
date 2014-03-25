package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type handlerFunc func([]string) (string, handlerFunc, []string)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print(">")
	line, isPrefix, err := reader.ReadLine()

	for err == nil && !isPrefix {
		parseLine(string(line))
		fmt.Print(">")
		line, isPrefix, err = reader.ReadLine()
	}
	if isPrefix {
		fmt.Println("Buffer size too small")
	}
	if err != io.EOF {
		fmt.Println(err)
	}

	fmt.Printf("Terminated\n")
}

func parseLine(line string) {
	//fmt.Println(line)
	commandTokens := strings.Fields(line)
	ans, parse, tokens := startParse(commandTokens)
	for parse != nil && len(tokens) != 0 {
		ans, parse, tokens = parse(tokens)
		if ans != "" {
			fmt.Printf("%s\n", ans)
		}
	}
}

//Commands
//List * -- lists * (currently echo *)
//start
//	et id + params?
//	heatmap id etid=id color/col/c=red/blue , width/w/x=a , height/h/y=b , duration/dur/d=c , frequency/freq,f=d
//	region id etid=id positionx/posx/px=a , positiony,posy,py=b , shape=rectangle/rect/elipse , height/h=c width/w=d
//stop
//	et id
//	heatmap id
//	region id
//set
//	heatmap id  a=b , c=d ... -- setup heatmap
//	region id a=b , c=d ... -- setup region
func startParse(tokens []string) (string, handlerFunc, []string) {
	if len(tokens) > 1 {
		switch tokens[0] {
		case "list":
			return "", parseList, tokens[1:]
		case "set":
			return "", parseSet, tokens[1:]
		}
	}
	//default
	return parseError(tokens[0])
}

func parseError(token string) (string, handlerFunc, []string) {
	fmt.Println("Error parsing \"" + token + "\".")
	return "", startParse, nil
}

// Stuff you can list
func parseList(tokens []string) (string, handlerFunc, []string) {
	return tokens[0], startParse, tokens[1:]
}

func parseSet(tokens []string) (string, handlerFunc, []string) {
	if len(tokens) > 1 {
		switch tokens[0] {
		case "heatmap":
			return "", parseHeatMap, tokens[1:]
		case "region":
			return "", parseRegion, tokens[1:]
		}
	}
	//default
	return parseError(tokens[0])
}

func parseHeatMap(tokens []string) (string, handlerFunc, []string) {
	return parseError(tokens[0])
	//TODO make update message and go through tokens and update every field in the message
	// if comma is found more settings follow. else end and continue from start parsing.
	//Also send the message.
}

func parseRegion(tokens []string) (string, handlerFunc, []string) {
	return parseError(tokens[0])
	//TODO make update message, send it, then start from beginning.
}
