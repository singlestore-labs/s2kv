package s2kv

import (
	"bufio"
	"log"
	"net"
	"strings"

	"github.com/secmask/go-redisproto"
)

type Server struct {
	db *SingleStore
}

func NewServer(db *SingleStore) *Server {
	return &Server{db: db}
}

func (s *Server) ListenAndServe(port string) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error on accept: ", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	parser := redisproto.NewParser(conn)
	writer := redisproto.NewWriter(bufio.NewWriter(conn))

	var ew error

	for {
		command, err := parser.ReadCommand()

		if err != nil {
			_, ok := err.(*redisproto.ProtocolError)
			if ok {
				ew = writer.WriteError(err.Error())
			} else {
				log.Println(err, " closed connection to ", conn.RemoteAddr())
				break
			}
		} else {
			cmd := strings.ToUpper(string(command.Get(0)))
			handler, ok := CommandHandlers[cmd]
			if !ok {
				ew = writer.WriteError("command not supported")
			} else {
				ew = handler(s.db, writer, command)
			}
		}

		if command.IsLast() {
			writer.Flush()
		}
		if ew != nil {
			_ = writer.WriteError(ew.Error())
			writer.Flush()
			log.Printf("Error on `%s`: %s", commandString(command), ew)
			break
		}
	}
}
