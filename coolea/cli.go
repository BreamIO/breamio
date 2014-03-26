package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/maxnordlund/breamio/aioli"
	"github.com/maxnordlund/breamio/beenleigh"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

type handlerFunc func([]string) (string, handlerFunc, []string)

var client *Client
var running = true
var ip = flag.String("ip", "localhost:4041", "Specify ip amd port")

func main() {
	flag.Parse()
	conn, err := net.Dial("tcp", *ip)
	if err != nil {
		log.Println("Could not connect to server:", err)
		return
	}
	defer conn.Close()
	client = NewClient(conn)
	fmt.Println("Connected to", *ip)

	reader := bufio.NewReader(os.Stdin)

	fmt.Print(">")
	line, isPrefix, err := reader.ReadLine()

	//Read a lot of input
	for err == nil && !isPrefix {
		parseLine(string(line))
		if !running {
			break // If we are not supposed to be running break this loop
		}
		fmt.Print(">")
		line, isPrefix, err = reader.ReadLine()
	}
	if isPrefix {
		fmt.Println("Buffer size too small")
	}
	if err != io.EOF && err != nil {
		fmt.Println(err)
	}
	client.Wait()
	//Done
	fmt.Printf("Terminated\n")
}

//Parses the line given
func parseLine(line string) {
	//fmt.Println(line) //Debug
	commandTokens := strings.Fields(line)
	ans, parse, tokens := startParse(commandTokens)
	//Start itterative parsing for all tokens
	for parse != nil && len(tokens) > 0 {
		ans, parse, tokens = parse(tokens)
		if ans != "" {
			fmt.Printf("%s\n", ans)
		}
	}
}

//Some kind of parsetree over the available commands
//--------------------------------------------------
//Commands
//List * -- lists * (currently echo *)
//start
//	et id optionsstring
//	heatmap id etid=id color/col/c=red/blue , width/w/x=a , height/h/y=b , duration/dur/d=c , frequency/freq,f=d
//	region id etid=id positionx/posx/px=a , positiony,posy,py=b , shape=rectangle/rect/elipse , height/h=c width/w=d
//stop
//	et id
//	heatmap id
//	region id
//--------------------------------------------------
func startParse(tokens []string) (string, handlerFunc, []string) {
	if len(tokens) > 0 && tokens[0] == "exit" {
		running = false
		return "exit program", nil, nil
	}

	if len(tokens) > 1 {
		switch tokens[0] {
		case "list":
			return "", parseList, tokens[1:]
		case "start":
			return "", parseStart, tokens[1:]
		}
	}
	//default
	if len(tokens) == 0 {
		return "", startParse, tokens
	}
	return parseError(tokens[0])
}

//The error given if the parser fails
func parseError(token string) (string, handlerFunc, []string) {
	fmt.Println("Error parsing \"" + token + "\".")
	return "", startParse, nil
}

//Parse the subtree of list
func parseList(tokens []string) (string, handlerFunc, []string) {
	return tokens[0], startParse, tokens[1:]
}

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
	client.Send(aioli.ExtPkg{"new:tracker", 256, payload})
	return "Sent request to start new ET", startParse, tokens[2:]
}

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
	client.Send(aioli.ExtPkg{"tracker:shutdown", 256, payload})
	return "Sent request to stop new ET", startParse, tokens[1:]
}
