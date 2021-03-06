package main
import (
	"fmt"
	"os"
	"bufio"
	"strings"
	"net"
	"time"
)

type Node struct {
	ip   string
	port string
	receive bool 
	send bool
}
var sendBackFlag bool
var iniFlag bool
var sendIniMessFlag bool
var Iam Node
var Parent Node
var neighbors []Node
var multicast []string
var multicastString string
var isJoin string
var joinString string

func main() {
	isJoin = "false"
	joinString = ""
	filename := `configuration.conf`
	fmt.Println("Start..."  )
	// Find Iam,Initiator,Neighbors
	readFile(filename)
	fmt.Printf("I am %s:%s \niniFlag is: %t \nAll my neighbors are: %v \n" , Iam.ip , Iam.port , iniFlag , neighbors)
	go server(Iam)
	
	if checkNeighborServer(neighbors) {
		done := false
		for {
			if iniFlag && !sendIniMessFlag {
				fmt.Println("Start to send message from initiator: "  )
				iniMss := "&Iam="+Iam.ip+":"+Iam.port+"&multicast="+multicastString+"&join=false"+"&path="+joinString
				sendMssToAllNeighbors(iniMss)
				sendIniMessFlag = true
				fmt.Printf("All my neighbors are: %v \n" , neighbors)
			}else{
				time.Sleep(3000 * time.Millisecond)
				fmt.Print("." )
				if checkReceiveFromAll() {
					if iniFlag {
						fmt.Println("\nDone "  )
						done = true
					}else{
						// check if join or not
						if isJoin == "false" {
							for i:=0; i < len(multicast);i++{
								if multicast[i] == Iam.port {
									isJoin = "true"
									fmt.Println("I Join to the path..."  )
									joinString = joinString + "," +Iam.port
								}
							}	
						}
						
						m := "&Iam="+Iam.ip+":"+Iam.port+"&multicast="+multicastString+"&join="+isJoin+"&path="+joinString
						sendMessage(m,Parent)
						fmt.Println("\nDone "  )
						done = true
					}
				}
			}
			if done {
				fmt.Println( neighbors , isJoin , joinString)
				break
			}
		}
	}
	
}

func readFile(fileName string){
	f, _ := os.Open(fileName)
	defer f.Close()
	r := bufio.NewReaderSize(f, 2*1024)
	line, isPrefix, err := r.ReadLine()
	i := 1
	for err == nil && !isPrefix {
		s := string(line)
		if i == 1 {
				// Find Iam
				t :=strings.Split(s, ":")
				Iam = Node{t[0],t[1],false,false}
		}else{
			k :=strings.Split(s, ":")
			if k[0] == "initiator" {
					// Find if initiator
					iniFlag = true
					// Find multicast nodes
					mNodes :=strings.Split(k[1], ",")
					for i:=0; i < len(mNodes);i++{
						multicast = append(multicast, mNodes[i])
					}
					multicastString=k[1]
					fmt.Println("Multicast Nodes: ",multicast)
					fmt.Println("Multicast String: ",multicastString)
				}else{
					// Find neighbors
					neighbors = append(neighbors, Node{k[0],k[1],false,false})
				}		
		}
		i++
		line, isPrefix, err = r.ReadLine()		
	}
}

func analizMessage(message string) map[string]string{
	ms :=strings.Split(message, "&")
	msIam :=strings.Split(ms[1], "=")
	mx :=strings.Split(msIam[1], ":")
	GetMessage := make(map[string]string)
	GetMessage["ip"] =  mx[0]
	GetMessage["port"] =  mx[1]
	
	getmulticast :=strings.Split(ms[2], "=")
	GetMessage["multicast"] =  getmulticast[1]
	mNodes :=strings.Split(getmulticast[1], ",")
	for i:=0; i < len(mNodes);i++{
		multicast = append(multicast, mNodes[i])
	}
	
	getisjoin :=strings.Split(ms[3], "=")
	GetMessage["join"] =  getisjoin[1]
	
	getpath :=strings.Split(ms[4], "=")
	GetMessage["path"] =  getpath[1]
					
	fmt.Println("GetMessage-> ",GetMessage)
	return GetMessage
}

func server(s Node) {
	fmt.Printf("Launching server... %s:%s \n" , s.ip,s.port)
	ln, _ := net.Listen("tcp", s.ip+":"+s.port)
	defer ln.Close()
	for {
		conn, _ := ln.Accept()
		message, _ := bufio.NewReader(conn).ReadString('\n') 
		if string(message) != "" {
			fmt.Println("->", string(message))
			doIt(analizMessage(message))
		}
	}	
	
}

func checkNeighborServer(n []Node) bool{
	for i:=0; i < len(n);i++{
		for {
			conn, err := net.Dial("tcp", n[i].ip+":"+n[i].port)
			fmt.Println("Looking for " + n[i].ip+":"+n[i].port)
			time.Sleep(3000 * time.Millisecond)
			if err == nil {
				conn.Close()
				break
			}
		}
	}
	
	return true
}

func sendMessage(s string, n Node){
	conn, _ := net.Dial("tcp", n.ip+":"+n.port)
	defer conn.Close()
	conn.Write([]byte(s))
	fmt.Printf("Message Sent to %s:%s \n" ,n.ip,n.port )	
}

func sendMssToAllNeighbors(ms string){

	for i:=0; i < len(neighbors);i++{
		if Parent.port != neighbors[i].port {
			sendMessage(ms,neighbors[i])
			neighbors[i].send = true	
		}
		
	}
}

func doIt( ms map[string]string){
	Sender,id := findNodeBtwNeighbors(ms["ip"],ms["port"])
	joinString = ms["path"]
	if isJoin == "false" {
		if ms["join"] == "true" {
			isJoin = "true"
			fmt.Println("I Join to the path..."  )
			joinString = joinString + "," +Iam.port
		}
	}
	
	if iniFlag{
		neighbors[id].receive = true
	}else{
		if Parent.ip == "" {
			Parent =  Sender
			multicastString = ms["multicast"]
			
			if isJoin == "true" {
				joinString = joinString + "," +Iam.port
			}
			
			sendMss := "&Iam="+Iam.ip+":"+Iam.port+"&multicast="+multicastString+"&join="+isJoin+"&path="+joinString
			neighbors[id].receive = true
			sendMssToAllNeighbors(sendMss)
		}else{
			neighbors[id].receive = true
		}
	}
	
}

func findNodeBtwNeighbors(ip string, port string) (Node , int){
	j := 0
	for i:=0; i < len(neighbors);i++{
		if neighbors[i].ip == ip && neighbors[i].port == port {
				j = i
		}
	}
	return neighbors[j],j
}

func checkReceiveFromAll() bool{
	Myreturn := true
	for _, n := range neighbors {
		if Parent.port != n.port   {
			if !n.receive {
				Myreturn = false
			}
		}
	}
	return Myreturn
}
