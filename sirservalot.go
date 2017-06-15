package main

import (
    "flag"
    "log"

    "./bridge"
)

func main() {
    var serialPort string
    var tcpBindAddress string

    flag.StringVar(&serialPort, "serial", "/dev/ttyUSB0", "Serial port to share.")
    flag.StringVar(&tcpBindAddress, "listen", ":1812", "Address:Port to listen for TCP on.")
    flag.Parse()

    log.Printf("Share %s to %s", serialPort, tcpBindAddress)

    serialReceived := make(chan []byte, 100)
    tcpReceived := make(chan []byte, 100)

    tcpServer := bridge.NewTCPServer(tcpBindAddress, serialReceived, tcpReceived)
    go tcpServer.Listen()

    bridge.Serial(serialPort, serialReceived, tcpReceived)
}
