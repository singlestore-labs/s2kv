package s2redis

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

outer:
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
			switch cmd {
			case "GET":
				key := string(command.Get(1))
				val, err := s.db.BlobGet(key)
				if err != nil {
					log.Printf("Error on GET %s: %s", key, err)
					break outer
				}
				if val == "" {
					ew = writer.WriteBulk(nil)
				} else {
					ew = writer.WriteBulkString(val)
				}
			case "DEL":
				key := string(command.Get(1))
				val, err := s.db.KeyDelete(key)
				if err != nil {
					log.Printf("Error on DEL %s: %s", key, err)
					break outer
				}
				if val {
					writer.WriteInt(1)
				} else {
					writer.WriteInt(0)
				}
			case "SET":
				key := string(command.Get(1))
				val := string(command.Get(2))
				err := s.db.BlobSet(key, val)
				if err != nil {
					log.Printf("Error on SET %s %s: %s", key, val, err)
					break outer
				}
				ew = writer.WriteBulkString("OK")
			default:
				ew = writer.WriteError("Command not support")
			}
		}
		if command.IsLast() {
			writer.Flush()
		}
		if ew != nil {
			log.Println("Connection closed", ew)
			break
		}
	}
}
