package server

import (
	"encoding/binary"
	"fmt"
	"github.com/atlasgurus/rulestone/actors"
	"github.com/atlasgurus/rulestone/condition"
	"io"
	"net"
	"unsafe"
)

type RulestoneServer struct {
	BaseRulestoneServer
	conn net.Conn
}

func NewRulestoneServer(conn net.Conn) *RulestoneServer {
	result := RulestoneServer{
		BaseRulestoneServer{
			base:        &RulestoneServer{conn: conn},
			ruleEngines: make([]*RuleEngineInfo, 0),
			matchActor:  actors.NewActor(nil, 10000),
		},
		conn,
	}
	return &result
}

const SocketPath = "/tmp/rulestone_sidecar.sock"
const SendBuferSize = 4 * 1024 * 1024
const RecvBuferSize = 4 * 1024 * 1024
const CreateRuleEngine = 1
const AddRuleFromJsonStringCommand = 2
const AddRuleFromYamlStringCommand = 3
const AddRulesFromFileCommand = 4
const AddRulesFromDirectoryCommand = 5
const ActivateCommand = 6
const MatchCommand = 7

//type MyRulestoneServiceServer struct {
//	grpc.UnimplementedRulestoneServiceServer
//	rs := NewRulestoneServer()
//}
//
//func (MyRulestoneServiceServer) CreateRuleEngine(context.Context, *grpc.EmptyRequest) (*grpc.RuleEngineResponse, error) {
//	return nil, status.Errorf(codes.Unimplemented, "method CreateRuleEngine not implemented")
//}

func (rs *RulestoneServer) HandleConnection() {
	defer rs.conn.Close()

	unixConn, ok := rs.conn.(*net.UnixConn)
	if !ok {
		fmt.Println("Connection is not a Unix domain socket connection")
		return
	}

	// Set the send and receive buffer sizes
	err := unixConn.SetWriteBuffer(SendBuferSize)
	if err != nil {
		fmt.Println("Error setting write buffer size:", err)
	}

	err = unixConn.SetReadBuffer(RecvBuferSize)
	if err != nil {
		fmt.Println("Error setting read buffer size:", err)
	}

	for {
		command, err := readInt16(rs.conn)
		if err != nil {
			if err == io.EOF {
				// Connection was closed by the client
				fmt.Println("Connection was closed by the client:", err)
				return
			}
			// Handle other errors, for example by logging them
			fmt.Println("Error reading message:", err)
			return
		}

		// Handle the command
		switch command {
		case CreateRuleEngine:
			ruleEngineID := rs.CreateNewRuleEngine()
			writeInt16(rs.conn, uint16(ruleEngineID))
		case ActivateCommand:
			ruleEngineID, err := readInt16(rs.conn)
			if err != nil {
				// Handle other errors, for example by logging them
				fmt.Println("Error reading message:", err)
				return
			}
			rs.ActivateRuleEngine(int(ruleEngineID))
		case AddRuleFromJsonStringCommand:
			ruleEngineID, err := readInt16(rs.conn)
			if err != nil {
				// Handle other errors, for example by logging them
				fmt.Println("Error reading message:", err)
				return
			}
			ruleString, err := readLengthPrefixedMessage(rs.conn)
			if err != nil {
				// Handle other errors, for example by logging them
				fmt.Println("Error reading message:", err)
				return
			}
			ruleId := rs.AddRuleFromString(int(ruleEngineID), ruleString, "json")
			writeInt32(rs.conn, uint32(ruleId))
		case AddRuleFromYamlStringCommand:
			ruleEngineID, err := readInt16(rs.conn)
			if err != nil {
				// Handle other errors, for example by logging them
				fmt.Println("Error reading message:", err)
				return
			}
			ruleString, err := readLengthPrefixedMessage(rs.conn)
			if err != nil {
				// Handle other errors, for example by logging them
				fmt.Println("Error reading message:", err)
				return
			}
			ruleId := rs.AddRuleFromString(int(ruleEngineID), ruleString, "yaml")
			writeInt32(rs.conn, uint32(ruleId))
		case AddRulesFromDirectoryCommand:
			ruleEngineID, err := readInt16(rs.conn)
			if err != nil {
				// Handle other errors, for example by logging them
				fmt.Println("Error reading message:", err)
				return
			}
			rulePath, err := readLengthPrefixedMessage(rs.conn)
			if err != nil {
				// Handle other errors, for example by logging them
				fmt.Println("Error reading message:", err)
				return
			}
			numRules := rs.AddRulesFromDirectory(int(ruleEngineID), rulePath)
			writeInt32(rs.conn, uint32(numRules))
		case AddRulesFromFileCommand:
			ruleEngineID, err := readInt16(rs.conn)
			if err != nil {
				// Handle other errors, for example by logging them
				fmt.Println("Error reading message:", err)
				return
			}
			rulePath, err := readLengthPrefixedMessage(rs.conn)
			if err != nil {
				// Handle other errors, for example by logging them
				fmt.Println("Error reading message:", err)
				return
			}
			numRules := rs.AddRulesFromFile(int(ruleEngineID), rulePath)
			writeInt32(rs.conn, uint32(numRules))
		case MatchCommand:
			if rs.processMatchCommand() != nil {
				return
			}
		}
	}
}

func (rs *RulestoneServer) processMatchCommand() error {
	ruleEngineID, err := readInt16(rs.conn)
	if err != nil {
		// Handle other errors, for example by logging them
		fmt.Println("Error reading message:", err)
		return err
	}
	/*
		requestId, err := readInt32(conn)
		if err != nil {
			// Handle other errors, for example by logging them
			fmt.Println("Error reading message:", err)
			return
		}

	*/
	jsonData, err := readLengthPrefixedMessage(rs.conn)
	return rs.PerformMatch(0, int(ruleEngineID), jsonData, nil)
}

func (rs *RulestoneServer) responseMatches(_ int64, matches []condition.RuleIdType, commObj interface{}) {
	writeMatchesList(rs.conn, matches)
}

func readInt32(conn net.Conn) (int32, error) {
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(conn, lengthBuf)
	if err != nil {
		return -1, err
	}
	return int32(binary.BigEndian.Uint32(lengthBuf)), nil
}

func readInt16(conn net.Conn) (int16, error) {
	v := make([]byte, 2)
	_, err := io.ReadFull(conn, v)
	if err != nil {
		return -1, err
	}
	return int16(binary.BigEndian.Uint16(v)), nil
}

// readLengthPrefixedMessage reads a length-prefixed string message from the connection.
func readLengthPrefixedMessage(conn net.Conn) (string, error) {
	length, err := readInt32(conn)
	if err != nil {
		return "", err
	}

	data := make([]byte, length)
	_, err = io.ReadFull(conn, data)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func writeInt32(conn net.Conn, v uint32) {
	buf32 := []byte{0, 0, 0, 0}
	binary.BigEndian.PutUint32(buf32, v)
	conn.Write(buf32)
}

func writeInt16(conn net.Conn, v uint16) {
	buf16 := []byte{0, 0}
	binary.BigEndian.PutUint16(buf16, v)
	conn.Write(buf16)
}

// writeLengthPrefixedMessage writes a length-prefixed string message to the connection.
func writeMatchesList(conn net.Conn, matches []condition.RuleIdType) {
	numMatches := len(matches)

	writeInt32(conn, uint32(len(matches)))

	if numMatches > 0 {
		byteSlice := (*[1 << 30]byte)(unsafe.Pointer(&matches[0]))[:len(matches)*4]
		conn.Write(byteSlice)
	}
}
