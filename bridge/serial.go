package bridge

import (
    "bufio"
    "errors"
    "io"
    "log"
    "os"
    "syscall"
)

// #include <termios.h>
// #include <unistd.h>
import "C"

func serialOpen(port string) (io.ReadWriteCloser, error) {
    file, err :=
        os.OpenFile(
            port,
            syscall.O_RDWR|syscall.O_NOCTTY,
            0600)
    if err != nil {
        return nil, err
    }

    fd := C.int(file.Fd())
    if C.isatty(fd) == 0 {
        err := errors.New("File is not a serial port")
        return nil, err
    }

    var termios C.struct_termios
    _, err = C.tcgetattr(fd, &termios)
    if err != nil {
        return nil, err
    }

    var baud C.speed_t
    baud = C.B115200
    _, err = C.cfsetispeed(&termios, baud)
    if err != nil {
        return nil, err
    }
    _, err = C.cfsetospeed(&termios, baud)
    if err != nil {
        return nil, err
    }
    return file, nil
}

func serialReader(port io.ReadWriteCloser, serialReceived chan []byte) {
    scanner := bufio.NewScanner(port)
    for {
        if scanner.Scan() {
            msg := scanner.Bytes()
            log.Printf("Read %d bytes from serial port", len(msg))
            serialReceived <- msg
        }
        if err := scanner.Err(); err != nil {
            log.Fatal(err)
        }
    }
}

func Serial(serialPort string, serialReceived chan []byte, tcpReceived chan []byte) {
    port, err := serialOpen(serialPort)
    if err != nil {
        log.Fatal(err)
    }
    defer port.Close()

    go serialReader(port, serialReceived)

    for {
        select {
        case msg := <-tcpReceived:
            log.Printf("Writing %d bytes to serial port", len(msg))
            port.Write(msg)
        }
    }
}
