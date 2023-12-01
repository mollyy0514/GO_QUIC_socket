package main

import (
	"fmt"
	"time"
)

// "flag"
// "strings"

func main() {
	// _ports := flag.String("p", "", "ports to bind (space-separated)")
	// // _hosts := flag.String("H", "140.112.20.183", "server ip address")
	// flag.Parse()

	// // ports := *_ports
	// portsList := strings.Split(*_ports, ",")

	// print("ports:", portsList[0], "&", portsList[1])
	// // print("HOST:", *_hosts)

	// devicesList := []string{"sm00", "sm01"}
	// for i := 0; i < len(devicesList); i++ {
	// 	print(devicesList[i])
	// }
	// portsList := Get_Port(devicesList)
	// print(portsList[0][0], "&", portsList[0][1])
	// print(portsList[1][0], "&", portsList[1][1])

	currentTime := time.Now()
	y := currentTime.Year()
	m := currentTime.Month()
	d := currentTime.Day()
	fmt.Printf("./data/capturequic_c_%02d%02d%02d.pcap", y, m, d)
}

// func Get_Port (devicesList []string) [][2]int {
// 	var portsList [][2]int
// 	for _, device := range devicesList {
// 		// default uplink port and downlink port for each device
// 		ports := []int{devices.Device_to_port[device][0], devices.Device_to_port[device][1]}
// 		portsList = append(portsList, [2]int(ports))
// 	}
// 	return portsList
// }
