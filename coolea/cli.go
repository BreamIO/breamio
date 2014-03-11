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

func startParse(tokens []string) (string, handlerFunc, []string) {
	if len(tokens) > 0 {
		switch tokens[0] {
		case "list":
			if len(tokens) > 1 {
				return "", parseList, tokens[1:]
			}
		}
	}
	//default
	fmt.Println("could not parse command\n")
	return "", startParse, nil
}

// Stuff you can list
func parseList(tokens []string) (string, handlerFunc, []string) {
	return tokens[0], startParse, tokens[1:]
}
