# sirservalot
This is at some stage going to be a serial server for linux.

The idea is that we can share a serial port to multiple network subscribers.

And it is in GO just for giggles, wanted to try the channels and go routines to see what a simple server was like. Due to Go's cross platform nature it is a little tricky playing with hardware, not quite io monad but...

## Simple Test Setup

Create a virtial serial device:
```
socat -d -d pty,raw,echo=0,link=/tmp/ttyS00 pty,raw,echo=0,link=/tmp/ttyS01
```

Run sirservalot:
```
go run sirservalot.go  -serial /tmp/ttyS00 -listen :1812
```

To provide input from serial device:
```
echo "done the thing" > /tmp/ttyS01
```

To read input from tcp client:
```
cat /tmp/ttyS01
```

Create tcp client (read and write):
```
nc localhost 1812
```
