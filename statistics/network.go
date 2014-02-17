package main

import (
	"net"
	"encoding/binary"
	"bytes"
	"time"
	"fmt"
	"os"
)

type CoordinateListener interface {
	Add(coord Coordinate)
}

type CoordinateManager struct {
	listeners []CoordinateListener
}

func newCoordinateManager() *CoordinateManager {
	return &CoordinateManager{
		listeners: make([]CoordinateListener, 0),
	}
}

func (cm *CoordinateManager) AddListener(listener CoordinateListener) {
	cm.listeners = append(cm.listeners, listener)
}

func (cm CoordinateManager) addCoord(c Coordinate) {
	for _, l := range cm.listeners {
		l.Add(c)
	}
}

func connectTo(ip string) *CoordinateManager {
	conn, err := net.Dial("tcp", ip)

	if err != nil {
		panic(err.Error())
	}

	buf := make([]byte, 0)

	buf = append(buf, []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}...)	
	
	conn.Write(buf)

	coordManager := newCoordinateManager()
	
	go func() {

		for {
			coord := getPoint(conn)
			if coord == nil {
				continue
			}

			coordManager.addCoord(*coord)
			fmt.Fprintln(os.Stderr, coord)
		}
	}()

	return coordManager
}

//Reads a point and drops all other messages
func getPoint(conn net.Conn) *Coordinate {
	var caseInt float64
	buf := make([]byte, 1)
	_, err := conn.Read(buf)
	if err != nil {
		panic(err.Error())
	}
	//buf.Append
	buf2 := bytes.NewBuffer(buf)
	fmt.Fprintln(os.Stderr, buf2)
	err = binary.Read(buf2, binary.BigEndian, &caseInt)

	if err != nil {
		panic(err.Error())
	}


	fmt.Fprintln(os.Stderr, caseInt)
	switch caseInt {
		case 1: return getPointHelper(conn)
		case 2:
			dropMessage(conn, 8)
			return nil
		case 3: return nil
		case 4: return nil
		case 5:	
			dropMessage(conn, 16)
			return nil
		case 6: return nil
		case 7: 
			dropMessage(conn, 8)
			return nil
		default: panic("We are not getting anything that is specified by protocol.")
	}
		
}
func dropMessage(conn net.Conn, numOfBytes int) *Coordinate{
	bufRead := make([]byte, numOfBytes)
	_, err := conn.Read(bufRead)
	if err != nil {
		panic(err.Error())
	}
	return nil
}

func getPointHelper(conn net.Conn) *Coordinate {
	var x, y float64
	time := time.Now()
	buf := make([]byte, 8)
	_, err := conn.Read(buf)
	if err != nil {
		panic(err.Error())
	}

	binary.Read(bytes.NewBuffer(buf), binary.BigEndian, &x)

	_, err = conn.Read(buf)
	if err != nil {
		panic(err.Error())
	}
	
	binary.Read(bytes.NewBuffer(buf), binary.BigEndian, &y)

	_, err = conn.Read(buf)
	if err != nil {
		panic(err.Error())
	}
	
	binary.Read(bytes.NewBuffer(buf), binary.BigEndian, nil)
	
	return NewCoordinate(x,y,time)
}








