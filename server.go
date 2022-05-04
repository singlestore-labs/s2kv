package s2redis

import (
	"bufio"
	"log"
	"net"
	"strconv"
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
			case "SET":
				key := string(command.Get(1))
				val := command.Get(2)
				err := s.db.BlobSet(key, val)
				if err != nil {
					log.Printf("Error on SET %s %s: %s", key, val, err)
					break outer
				}
				ew = writer.WriteBulkString("OK")
			case "GET":
				key := string(command.Get(1))
				val, err := s.db.BlobGet(key)
				if err != nil {
					log.Printf("Error on GET %s: %s", key, err)
					break outer
				}
				ew = writer.WriteBulk(val)
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
			case "FLUSHALL":
				err := s.db.FlushAll()
				if err != nil {
					log.Printf("Error on FLUSHALL: %s", err)
					break outer
				}
				ew = writer.WriteBulkString("OK")
			case "KEYS":
				pattern := string(command.Get(1))
				if pattern == "" {
					pattern = "%"
				}
				out, err := s.db.Keys(pattern)
				if err != nil {
					log.Printf("Error on KEYS: %s", err)
					break outer
				}
				ew = writer.WriteBulksSlice(out)
			case "EXISTS":
				key := string(command.Get(1))
				exists, err := s.db.KeyExists(key)
				if err != nil {
					log.Printf("Error on EXISTS %s: %s", key, err)
					break outer
				}
				if exists {
					ew = writer.WriteInt(1)
				} else {
					ew = writer.WriteInt(0)
				}

			// list functions
			case "RPUSH":
				key := string(command.Get(1))
				val := command.Get(2)
				err := s.db.ListAppend(key, val)
				if err != nil {
					log.Printf("Error on RPUSH %s %s: %s", key, val, err)
					break outer
				}
				ew = writer.WriteBulkString("OK")
			case "LREM":
				key := string(command.Get(1))
				val := command.Get(2)
				n, err := s.db.ListRemove(key, val)
				if err != nil {
					log.Printf("Error on RPUSH %s %s: %s", key, val, err)
					break outer
				}
				ew = writer.WriteInt(n)
			case "LRANGE":
				key := string(command.Get(1))
				start, err := strconv.Atoi(string(command.Get(2)))
				if err != nil {
					log.Printf("Error on LRANGE %s: %s", key, err)
					break outer
				}
				stop, err := strconv.Atoi(string(command.Get(3)))
				if err != nil {
					log.Printf("Error on LRANGE %s: %s", key, err)
					break outer
				}

				var out [][]byte
				if start == 0 && stop == -1 {
					out, err = s.db.ListGet(key)
					if err != nil {
						log.Printf("Error on LRANGE %s %d %d: %s", key, start, stop, err)
						break outer
					}
				} else {
					out, err = s.db.ListRange(key, start, stop)
					if err != nil {
						log.Printf("Error on LRANGE %s %d %d: %s", key, start, stop, err)
						break outer
					}
				}
				ew = writer.WriteBulksSlice(out)

			// set functions
			case "SADD":
				key := string(command.Get(1))
				val := command.Get(2)
				err := s.db.SetAdd(key, val)
				if err != nil {
					log.Printf("Error on SADD %s %s: %s", key, val, err)
					break outer
				}
				ew = writer.WriteBulkString("OK")
			case "SREM":
				key := string(command.Get(1))
				val := command.Get(2)
				n, err := s.db.SetRemove(key, val)
				if err != nil {
					log.Printf("Error on SREM %s %s: %s", key, val, err)
					break outer
				}
				ew = writer.WriteInt(n)
			case "SMEMBERS":
				key := string(command.Get(1))
				out, err := s.db.SetGet(key)
				if err != nil {
					log.Printf("Error on SMEMBERS %s: %s", key, err)
					break outer
				}
				ew = writer.WriteBulksSlice(out)
			case "SUNION":
				keys := make([]string, 0, command.ArgCount()-1)
				for i := 1; i < command.ArgCount(); i++ {
					keys = append(keys, string(command.Get(i)))
				}
				out, err := s.db.SetUnion(keys...)
				if err != nil {
					log.Printf("Error on SUNION %v: %s", keys, err)
					break outer
				}
				ew = writer.WriteBulksSlice(out)
			case "SINTER":
				keys := make([]string, 0, command.ArgCount()-1)
				for i := 1; i < command.ArgCount(); i++ {
					keys = append(keys, string(command.Get(i)))
				}
				out, err := s.db.SetIntersect(keys...)
				if err != nil {
					log.Printf("Error on SINTER %v: %s", keys, err)
					break outer
				}
				ew = writer.WriteBulksSlice(out)

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
