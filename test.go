package main

import (
	"flag"
	"strings"
)

func main() {
	_ports := flag.String("p", "", "ports to bind (space-separated)")
	// _hosts := flag.String("H", "140.112.20.183", "server ip address")
	flag.Parse()

	// ports := *_ports
	portsList := strings.Split(*_ports, ",")

	print("ports:", portsList[0], "&", portsList[1])
	// print("HOST:", *_hosts)
}
