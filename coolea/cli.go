package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/maxnordlund/breamio/aioli"
	. "github.com/maxnordlund/breamio/aioli/client"
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
var ip = flag.String("ip", "localhost:4041", "Specify ip and port. e.g. 127.0.0.1:4041 or localhost:4041.")

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
		case "stop":
			return "", parseStop, tokens[1:]
		case "create":
			return "", parseCreate, tokens[1:]
		case "update":
			return "", parseUpdate, tokens[1:]
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

func str2float(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return parseError(s)
	}
	return f
	payload, err := json.Marshal(struct{}{})
	client.Send(aioli.ExtPkg{"tracker:shutdown", false, id, payload, nil})
	return "Sent request to stop ET " + tokens[0], startParse, tokens[1:]
}
