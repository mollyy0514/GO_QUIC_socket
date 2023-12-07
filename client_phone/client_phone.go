package main

import (
	// "bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"GO_QUIC_socket/devices"
)

func main() {
	// Define command-line flags
	// _password := flag.String("p", "", "password for sudo command")
	_host := flag.String("H", "192.168.1.79", "server ip address")
	_devices := flag.String("d", "sm00", "list of devices (space-separated)")
	// _ports := flag.String("p", "3200", "ports to bind (space-separated)")
	_bitrate := flag.String("b", "1M", "target bitrate in bits/sec (0 for unlimited)")
	_length := flag.String("l", "250", "length of buffer to read or write in bytes (packet size)")
	_duration := flag.Int("t", 3600, "time in seconds to transmit for (default 1 hour = 3600 secs)")
	// Parse command-line arguments
	flag.Parse()

	_devices_string := *_devices
	devicesList, serialsList := Get_Devices_and_Serials(_devices_string)

	portsList := Get_Ports(devicesList)

	for i := 0; i < len(portsList); i++ {
		fmt.Printf("device: %s %s %d \n", devicesList[i], serialsList[i], portsList[i][0])
	}

	var wg sync.WaitGroup
	wg.Add(len(portsList))
	for i := 0; i < len(portsList); i++ {
		fmt.Printf("device: %s start \n", devicesList[i])
		port := fmt.Sprintf("%d,%d", portsList[i][0], portsList[i][1])
		defer wg.Done()
		// command := fmt.Sprintf("echo %s | sudo -S go run ./socket/client_socket_phone.go -H %s -d %s -p %s -b %s -l %s -t %d", *_password, *_host, *_devices, port, *_bitrate, *_length, *_duration)
		command := fmt.Sprintf("adb -s %s shell su -c cd /data/data/com.termux/files/home/GO_QUIC_socket && go run ./socket/client_socket_phone.go -H %s -d %s -p %s -b %s -l %s -t %d", serialsList[i], *_host, *_devices, port, *_bitrate, *_length, *_duration)
		fmt.Print(command)
		cmd := exec.Command("sh", "-c", command)

		// Set the working directory for the command
		cmd.Dir = "/data/data/com.termux/files/home/GO_QUIC_socket"

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		// err := cmd.Run()	// kill the program when times up, but won't kill tcpdump
		err := cmd.Start()	// won't kill the program when times up 
		if err != nil {
			// fmt.Println(fmt.Sprint(err) + ": " + cmd.Stderr)
			// return
			fmt.Print("err: ")
			fmt.Println(err)
			return
		}
		// Print the combined output
		fmt.Println(cmd.Stdout)
		// Socket(_host, _devices, &port, _bitrate, _length, _duration)
	}

	print("---End Of File---")
}

func Get_Devices_and_Serials(_devices_string string) ([]string, []string) {
	var devicesList []string
	var serialsList []string
	if strings.Contains(_devices_string, "-") {
		pmodel := _devices_string[:2]
		start, _ := strconv.Atoi(_devices_string[2:4])
		stop, _ := strconv.Atoi(_devices_string[5:7])
		for i := start; i <= stop; i++ {
			_dev := fmt.Sprintf("%s%02d", pmodel, i)
			devicesList = append(devicesList, _dev)
			serial := devices.Device_to_serial[_dev]
			serialsList = append(serialsList, serial)
		}
	} else {
		devicesList = strings.Split(_devices_string, " ")
		for _, dev := range devicesList {
			serial := devices.Device_to_serial[dev]
			serialsList = append(serialsList, serial)
		}
	}

	return devicesList, serialsList
}

func Get_Ports(devicesList []string) [][2]int {
	var portsList [][2]int
	for _, device := range devicesList {
		// default uplink port and downlink port for each device
		ports := []int{devices.Device_to_port[device][0], devices.Device_to_port[device][1]}
		portsList = append(portsList, [2]int(ports))
	}
	return portsList
}
