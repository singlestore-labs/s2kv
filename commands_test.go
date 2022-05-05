package s2kv_test

import (
	"flag"
	"fmt"
	"s2kv"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

var flagConfigPath = flag.String("config", "config.example.toml", "path to an optional config file")

func GetSingleStore(t *testing.T) *s2kv.SingleStore {
	configPath := *flagConfigPath
	config := s2kv.Config{}
	if configPath != "" {
		err := s2kv.LoadTOMLFiles(&config, []string{configPath})
		if err != nil {
			t.Fatal(err)
		}
	}

	db, err := s2kv.NewSingleStore(config.Database)
	if err != nil {
		t.Fatal(err)
	}

	return db
}

type GomegaMatcher struct {
	types.GomegaMatcher
}

func (m *GomegaMatcher) Matches(x interface{}) bool {
	out, _ := m.GomegaMatcher.Match(x)
	return out
}

func (m *GomegaMatcher) Got(x interface{}) string {
	return m.GomegaMatcher.FailureMessage(x)
}

func (m *GomegaMatcher) Want(x interface{}) string {
	return fmt.Sprintf("matches %v (%T)", x, x)
}

func (m *GomegaMatcher) String() string {
	return "matches"
}

func Match(matcher types.GomegaMatcher) gomock.Matcher {
	return &GomegaMatcher{matcher}
}

type TestOp struct {
	cmd   []string
	write func(writer *MockWriter) *gomock.Call
}

func mockCmd(name string, args ...string) TestOp {
	return TestOp{cmd: append([]string{name}, args...)}
}

func mockSimpleString(v string) TestOp {
	return TestOp{
		write: func(writer *MockWriter) *gomock.Call {
			return writer.EXPECT().WriteSimpleString(v)
		},
	}
}

func mockInt(v int) TestOp {
	return TestOp{
		write: func(writer *MockWriter) *gomock.Call {
			return writer.EXPECT().WriteInt(int64(v))
		},
	}
}

func mockBulk(v interface{}) TestOp {
	return TestOp{
		write: func(writer *MockWriter) *gomock.Call {
			if v == nil {
				return writer.EXPECT().WriteBulk(nil)
			} else {
				return writer.EXPECT().WriteBulk(Match(gomega.BeEquivalentTo(v)))
			}

		},
	}
}

func mockBulks(v ...string) TestOp {
	x := make([]interface{}, len(v))
	for i, s := range v {
		x[i] = []byte(s)
	}
	return TestOp{
		write: func(writer *MockWriter) *gomock.Call {
			return writer.EXPECT().WriteBulks(Match(gomega.ConsistOf(x)))
		},
	}
}

func TestAll(t *testing.T) {
	type test struct {
		name string
		ops  []TestOp
	}

	tests := []test{
		{
			name: "PING",
			ops: []TestOp{
				mockCmd("PING"),
				mockSimpleString("PONG"),
			},
		},
		{
			name: "SET then GET",
			ops: []TestOp{
				mockCmd("SET", "foo", "bar"),
				mockSimpleString("OK"),
				mockCmd("GET", "foo"),
				mockBulk("bar"),
			},
		},
		{
			name: "DEL",
			ops: []TestOp{
				mockCmd("SET", "key", "value"),
				mockSimpleString("OK"),
				mockCmd("GET", "key"),
				mockBulk("value"),
				mockCmd("DEL", "key"),
				mockInt(1),
				mockCmd("GET", "key"),
				mockBulk(nil),
			},
		},
		{
			name: "FLUSHALL",
			ops: []TestOp{
				mockCmd("SET", "key", "value"),
				mockSimpleString("OK"),
				mockCmd("FLUSHALL"),
				mockSimpleString("OK"),
				mockCmd("GET", "key"),
				mockBulk(nil),
			},
		},
		{
			name: "KEYS",
			ops: []TestOp{
				mockCmd("SET", "key", "value"),
				mockSimpleString("OK"),
				mockCmd("SET", "foo", "bar"),
				mockSimpleString("OK"),
				mockCmd("KEYS", ""),
				mockBulks("key", "foo"),
			},
		},
		{
			name: "EXISTS",
			ops: []TestOp{
				mockCmd("EXISTS", "key"),
				mockInt(0),
				mockCmd("SET", "key", "value"),
				mockSimpleString("OK"),
				mockCmd("EXISTS", "key"),
				mockInt(1),
			},
		},
		{
			name: "RPUSH",
			ops: []TestOp{
				mockCmd("RPUSH", "foo", "bar"),
				mockSimpleString("OK"),
				mockCmd("RPUSH", "foo", "baz"),
				mockSimpleString("OK"),
				mockCmd("LRANGE", "foo", "0", "-1"),
				mockBulks("bar", "baz"),
				mockCmd("RPUSH", "foo", "baz"),
				mockSimpleString("OK"),
				mockCmd("LRANGE", "foo", "0", "-1"),
				mockBulks("bar", "baz", "baz"),
			},
		},
	}

	db := GetSingleStore(t)

	for _, testConfig := range tests {
		t.Run(testConfig.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			writer := NewMockWriter(ctrl)

			// clear the db before each test
			err := db.FlushAll()
			if err != nil {
				t.Fatal(err)
			}

			var lastCall *gomock.Call
			var nextCall *gomock.Call

			cmds := make([]s2kv.Command, 0, len(testConfig.ops))
			for _, op := range testConfig.ops {
				if op.cmd != nil {
					cmds = append(cmds, NewCmd(ctrl, op.cmd...))
				} else if op.write != nil {
					nextCall = op.write(writer)
					if lastCall != nil {
						nextCall.After(lastCall)
					}
					lastCall = nextCall
				}
			}

			for _, cmd := range cmds {
				err := s2kv.CommandHandlers[string(cmd.Get(0))](db, writer, cmd)
				if err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func NewCmd(ctrl *gomock.Controller, args ...string) *MockCommand {
	cmd := NewMockCommand(ctrl)
	cmd.EXPECT().ArgCount().Return(len(args)).AnyTimes()
	for i, arg := range args {
		cmd.EXPECT().Get(i).Return([]byte(arg))
	}
	return cmd
}
