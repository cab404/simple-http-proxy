package main

import (
	"log"
	"net"
	"net/http"
	"strings"
)

type ServerContext struct {
	Routes        []Route
	ListenAddress string
}

type Route struct {
	ID   string
	Host string
}

func (context ServerContext) Find(seg string) *Route {

	for i := range context.Routes {
		if context.Routes[i].ID == seg {
			return &context.Routes[i]
		}
	}
	return nil
}

// Transfers contents of stream A to stream B
func Pipe(a net.Conn, b net.Conn) {
	inb := make([]byte, BUFFER_PIPE)
	defer a.Close()
	defer b.Close()
	for {
		n, err := a.Read(inb)
		if err != nil {
			break
		}
		_, err = b.Write(inb[:n])
		if err != nil {
			break
		}
	}
}

func (context ServerContext) ProcessConnection(client_socket net.Conn) {

	respond_with := func(code int) {
		resp := http.Response{Status: http.StatusText(code), StatusCode: code, ProtoMajor: 1}
		resp.Write(client_socket)
	}

	// Parsing HTTP status line
	// We don't want to get OOM-d, so our max status line length would be 10KB
	// Nobody needs longer than that, right?

	status_line := strings.Builder{}
	{
		status_max := BUFFER_STATUS_LINE
		buf := make([]byte, 1)
		status_line.Grow(status_max)
		for status_max > 0 {
			client_socket.Read(buf)
			status_line.Write(buf)
			if buf[0] == '\n' {
				break
			}
		}
		// If we hit the limit, client is probably hitting us with garbage.
		if status_max == 0 {
			log.Println("No line feed after ", status_max, " bytes, closing connection.")
			client_socket.Close()
			return
		}

	}

	// Now we'll verify and parse status line.
	split := strings.Split(status_line.String(), " ")
	if len(split) != 3 { // Status line is strange.
		respond_with(http.StatusBadRequest)
		client_socket.Close()
		return
	}

	method := split[0]
	path := split[1]
	protocol := split[2]

	split_path := strings.SplitN(path, "/", 3)

	if len(split_path) != 3 { // Path is malformed
		respond_with(http.StatusBadRequest)
		client_socket.Close()
		return
	}

	target_host := context.Find(split_path[1])
	modfied_path := "/" + split_path[2]

	if target_host == nil { // Host not found
		respond_with(http.StatusNotAcceptable)
		client_socket.Close()
		return
	}

	// We now have enough info to connect to target host and proxy request.
	target_socket, err := net.Dial("tcp", target_host.Host)

	if err != nil { // Failed to establish connection to target host
		respond_with(http.StatusBadGateway)
		client_socket.Close()
		return
	}

	// Since we read status line ourselves (and it should be modified),
	// we write it before launching duplex pipe
	target_socket.Write([]byte(method + " " + modfied_path + " " + protocol))

	// Letting them go
	go Pipe(client_socket, target_socket)
	go Pipe(target_socket, client_socket)

}

func main() {

	context := Config()
	f, _ := net.Listen("tcp", context.ListenAddress)

	for {
		a, err := f.Accept()
		log.Println("New connection from ", a.RemoteAddr())
		if err != nil {
			continue
		}
		go context.ProcessConnection(a)
	}

}
