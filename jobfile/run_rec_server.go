package jobfile

import (
	"fmt"
	"net"
	"time"

	"github.com/dshearer/jobber/common"
)

/*
An instance of this struct corresponds to a goroutine that is responsible for
sending run records over a certain connection.
*/
type connHandler struct {
	running    bool
	runRecChan chan []byte
	dropped    int
}

/*
Launch a goroutine that writes run records to the given connection.
When the goroutine exits, the connection will be closed.
*/
func launchConnHandler(conn net.Conn) *connHandler {
	handler := connHandler{
		running:    true,
		runRecChan: make(chan []byte, 10),
	}
	go handler.thread(conn)
	return &handler
}

func (self connHandler) Running() bool {
	return self.running
}

func (self *connHandler) thread(conn net.Conn) {
	defer func() {
		self.running = false
		conn.Close()
	}()

	for {
		select {
		case rec, ok := <-self.runRecChan:
			if !ok {
				return
			}

			// send run record
			timeout := time.Now().Add(5 * time.Minute)
			conn.SetWriteDeadline(timeout)
			if _, err := conn.Write(rec); err != nil {
				common.ErrLogger.Printf(err.Error())
				return
			}
		}
	} // for
}

func (self *connHandler) Push(runRec []byte) {
	if !self.running {
		return
	}

	select {
	case self.runRecChan <- runRec:
	default:
		self.dropped++
	}
}

func (self *connHandler) Stop() {
	if !self.running {
		return
	}
	close(self.runRecChan)
}

func acceptConns(listener net.Listener) <-chan net.Conn {
	connChan := make(chan net.Conn)

	go func() {
		defer close(connChan)
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			connChan <- conn
		}
	}()

	return connChan
}

type serverId struct {
	Proto   string
	Address string
}

/*
This is a TCP server listening on a particular port.
*/
type runRecServer struct {
	running    bool
	runRecChan chan []byte
	listener   net.Listener
	dropped    int
	handlers   []*connHandler
}

func launchRunRecServer(sid serverId) (*runRecServer, error) {
	// make socket
	listener, err := net.Listen(sid.Proto, sid.Address)
	if err != nil {
		return nil, err
	}

	// make server object
	server := runRecServer{
		running:    true,
		runRecChan: make(chan []byte, 10),
		listener:   listener,
	}

	// launch thread
	go server.thread()

	return &server, nil
}

func (self *runRecServer) Running() bool {
	return self.running
}

func (self *runRecServer) thread() {
	connChan := acceptConns(self.listener)

	defer func() {
		for _, handler := range self.handlers {
			handler.Stop()
		}
		self.handlers = nil
		self.running = false
	}()

	for {
		// reap dead handlers
		var newHandlers []*connHandler
		for _, handler := range self.handlers {
			if handler.Running() {
				newHandlers = append(newHandlers, handler)
			}
		}
		self.handlers = newHandlers

		// handle events
		select {
		case conn, ok := <-connChan:
			if !ok {
				/* listener died */
				return
			}

			// make handler for this connection
			self.handlers = append(self.handlers, launchConnHandler(conn))

		case runRec, ok := <-self.runRecChan:
			if !ok {
				return
			}

			// send run record to all connections
			for _, handler := range self.handlers {
				handler.Push(runRec)
			}
		} // select
	} // for
}

func (self *runRecServer) Stop() {
	if !self.running {
		return
	}

	self.listener.Close()
	close(self.runRecChan)
	self.running = false
}

func (self *runRecServer) Push(runRec []byte) {
	if !self.running {
		return
	}

	select {
	case self.runRecChan <- runRec:
	default:
		self.dropped++
	}
}

type runRecServerRegistry struct {
	servers map[serverId]*runRecServer
}

var GlobalRunRecServerRegistry runRecServerRegistry

func (self *runRecServerRegistry) SetServers(protos, addresses []string) {
	if len(protos) != len(addresses) {
		panic("len(protos) != len(addresses)")
	}

	if self.servers == nil {
		self.servers = make(map[serverId]*runRecServer)
	}

	sidMap := make(map[serverId]bool)
	for i := 0; i < len(protos); i++ {
		sid := serverId{Proto: protos[i], Address: addresses[i]}
		sidMap[sid] = true
	}

	var serversToMake []serverId
	for sid, _ := range sidMap {
		_, ok := self.servers[sid]
		if !ok {
			serversToMake = append(serversToMake, sid)
		}
	}

	var serversToDelete []serverId
	for sid, _ := range self.servers {
		_, ok := sidMap[sid]
		if !ok {
			serversToDelete = append(serversToDelete, sid)
		}
	}

	for _, sid := range serversToMake {
		server, err := launchRunRecServer(sid)
		if err != nil {
			common.ErrLogger.Println(err.Error())
			continue
		}
		self.servers[sid] = server
	}

	for _, sid := range serversToDelete {
		self.servers[sid].Stop()
		delete(self.servers, sid)
	}
}

func (self *runRecServerRegistry) Push(proto, address string, runRec []byte) {
	// get server
	sid := serverId{Proto: proto, Address: address}
	server, ok := self.servers[sid]
	if !ok {
		panic(fmt.Sprintf("No server for %v", sid))
	}
	if !server.Running() {
		return
	}

	// send run record
	server.Push(runRec)
}
