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

func mockBulkStrings(v ...string) TestOp {
	return TestOp{
		write: func(writer *MockWriter) *gomock.Call {
			return writer.EXPECT().WriteBulkStrings(Match(gomega.ConsistOf(v)))
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
			name: "GET",
			ops: []TestOp{
				mockCmd("SET", "foo", "bar"),
				mockSimpleString("OK"),
				mockCmd("GET", "foo"),
				mockBulk("bar"),
				mockCmd("SADD", "set", "foo"),
				mockSimpleString("OK"),
				mockCmd("GET", "set"),
				mockBulk(nil),
			},
		},
		{
			name: "SET",
			ops: []TestOp{
				mockCmd("SET", "foo", "bar"),
				mockSimpleString("OK"),
				mockCmd("GET", "foo"),
				mockBulk("bar"),
				mockCmd("SET", "foo", "baz"),
				mockSimpleString("OK"),
				mockCmd("GET", "foo"),
				mockBulk("baz"),
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
			name: "INCRBY",
			ops: []TestOp{
				mockCmd("INCRBY", "foo", "0"),
				mockInt(0),
				mockCmd("GET", "foo"),
				mockBulk("0"),
				mockCmd("INCRBY", "foo", "1"),
				mockInt(1),
				mockCmd("GET", "foo"),
				mockBulk("1"),
				mockCmd("INCRBY", "foo", "10"),
				mockInt(11),
				mockCmd("GET", "foo"),
				mockBulk("11"),
				mockCmd("INCRBY", "foo", "-5"),
				mockInt(6),
				mockCmd("GET", "foo"),
				mockBulk("6"),
				mockCmd("INCRBY", "bar", "-5"),
				mockInt(-5),
				mockCmd("GET", "bar"),
				mockBulk("-5"),
				mockCmd("INCRBY", "baz", "100"),
				mockInt(100),
				mockCmd("GET", "baz"),
				mockBulk("100"),
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
		{
			name: "LREM",
			ops: []TestOp{
				mockCmd("RPUSH", "foo", "bar"),
				mockSimpleString("OK"),
				mockCmd("RPUSH", "foo", "baz"),
				mockSimpleString("OK"),
				mockCmd("LREM", "foo", "bar"),
				mockInt(1),
				mockCmd("LRANGE", "foo", "0", "-1"),
				mockBulks("baz"),
				mockCmd("RPUSH", "foo", "baz"),
				mockSimpleString("OK"),
				mockCmd("LRANGE", "foo", "0", "-1"),
				mockBulks("baz", "baz"),
				mockCmd("LREM", "foo", "baz"),
				mockInt(2),
				mockCmd("LRANGE", "foo", "0", "-1"),
				mockBulks(),
			},
		},
		{
			name: "LRANGE",
			ops: []TestOp{
				mockCmd("RPUSH", "foo", "1"),
				mockSimpleString("OK"),
				mockCmd("RPUSH", "foo", "2"),
				mockSimpleString("OK"),
				mockCmd("RPUSH", "foo", "3"),
				mockSimpleString("OK"),
				mockCmd("LRANGE", "foo", "0", "-1"),
				mockBulks("1", "2", "3"),
				mockCmd("LRANGE", "foo", "0", "0"),
				mockBulks("1"),
				mockCmd("LRANGE", "foo", "1", "1"),
				mockBulks("2"),
				mockCmd("LRANGE", "foo", "0", "1"),
				mockBulks("1", "2"),
				mockCmd("LRANGE", "foo", "1", "2"),
				mockBulks("2", "3"),
				mockCmd("LRANGE", "foo", "0", "2"),
				mockBulks("1", "2", "3"),
				mockCmd("LRANGE", "foo", "0", "100"),
				mockBulks("1", "2", "3"),
			},
		},
		{
			name: "SADD",
			ops: []TestOp{
				mockCmd("SADD", "foo", "1"),
				mockSimpleString("OK"),
				mockCmd("SMEMBERS", "foo"),
				mockBulks("1"),
				mockCmd("SADD", "foo", "1"),
				mockSimpleString("OK"),
				mockCmd("SMEMBERS", "foo"),
				mockBulks("1"),
				mockCmd("SADD", "foo", "2"),
				mockSimpleString("OK"),
				mockCmd("SMEMBERS", "foo"),
				mockBulks("1", "2"),
			},
		},
		{
			name: "SREM",
			ops: []TestOp{
				mockCmd("SADD", "foo", "1"),
				mockSimpleString("OK"),
				mockCmd("SADD", "foo", "2"),
				mockSimpleString("OK"),
				mockCmd("SADD", "foo", "3"),
				mockSimpleString("OK"),
				mockCmd("SREM", "foo", "1"),
				mockInt(1),
				mockCmd("SMEMBERS", "foo"),
				mockBulks("2", "3"),
				mockCmd("SREM", "foo", "1"),
				mockInt(0),
				mockCmd("SMEMBERS", "foo"),
				mockBulks("2", "3"),
				mockCmd("SREM", "foo", "2"),
				mockInt(1),
				mockCmd("SREM", "foo", "3"),
				mockInt(1),
				mockCmd("SMEMBERS", "foo"),
				mockBulks(),
			},
		},
		{
			name: "SMEMBERS",
			ops: []TestOp{
				mockCmd("SADD", "foo", "1"),
				mockSimpleString("OK"),
				mockCmd("SADD", "foo", "2"),
				mockSimpleString("OK"),
				mockCmd("SADD", "foo", "3"),
				mockSimpleString("OK"),
				mockCmd("SMEMBERS", "foo"),
				mockBulks("1", "2", "3"),
			},
		},
		{
			name: "SINTER",
			ops: []TestOp{
				mockCmd("SADD", "foo", "1"),
				mockSimpleString("OK"),
				mockCmd("SADD", "foo", "2"),
				mockSimpleString("OK"),
				mockCmd("SADD", "foo", "3"),
				mockSimpleString("OK"),
				mockCmd("SADD", "bar", "3"),
				mockSimpleString("OK"),
				mockCmd("SADD", "bar", "4"),
				mockSimpleString("OK"),
				mockCmd("SADD", "bar", "5"),
				mockSimpleString("OK"),
				mockCmd("SADD", "baz", "5"),
				mockSimpleString("OK"),
				mockCmd("SADD", "baz", "6"),
				mockSimpleString("OK"),
				mockCmd("SADD", "baz", "7"),
				mockSimpleString("OK"),
				mockCmd("SINTER", "foo", "bar"),
				mockBulks("3"),
				mockCmd("SINTER", "foo", "bar", "baz"),
				mockBulks(),
				mockCmd("SADD", "baz", "3"),
				mockSimpleString("OK"),
				mockCmd("SINTER", "foo", "bar", "baz"),
				mockBulks("3"),
				mockCmd("SADD", "t", "1"),
				mockSimpleString("OK"),
				mockCmd("SADD", "t2", "2"),
				mockSimpleString("OK"),
				mockCmd("SINTER", "t", "t2"),
				mockBulks(),
			},
		},
		{
			name: "SUNION",
			ops: []TestOp{
				mockCmd("SADD", "foo", "1"),
				mockSimpleString("OK"),
				mockCmd("SADD", "foo", "2"),
				mockSimpleString("OK"),
				mockCmd("SADD", "foo", "3"),
				mockSimpleString("OK"),
				mockCmd("SADD", "bar", "3"),
				mockSimpleString("OK"),
				mockCmd("SADD", "bar", "4"),
				mockSimpleString("OK"),
				mockCmd("SADD", "bar", "5"),
				mockSimpleString("OK"),
				mockCmd("SADD", "baz", "5"),
				mockSimpleString("OK"),
				mockCmd("SADD", "baz", "6"),
				mockSimpleString("OK"),
				mockCmd("SUNION", "foo", "bar"),
				mockBulks("1", "2", "3", "4", "5"),
				mockCmd("SUNION", "foo", "bar", "baz"),
				mockBulks("1", "2", "3", "4", "5", "6"),
				mockCmd("SADD", "t", "1"),
				mockSimpleString("OK"),
				mockCmd("SADD", "t2", "2"),
				mockSimpleString("OK"),
				mockCmd("SUNION", "t", "t2"),
				mockBulks("1", "2"),
			},
		},
		{
			name: "SWITHMEMBER",
			ops: []TestOp{
				mockCmd("SADD", "foo", "1"),
				mockSimpleString("OK"),
				mockCmd("SADD", "baz", "1"),
				mockSimpleString("OK"),
				mockCmd("SADD", "baz", "2"),
				mockSimpleString("OK"),
				mockCmd("SADD", "bar", "1"),
				mockSimpleString("OK"),
				mockCmd("SADD", "bar", "2"),
				mockSimpleString("OK"),
				mockCmd("SWITHMEMBER", "1"),
				mockBulkStrings("foo", "bar", "baz"),
				mockCmd("SWITHMEMBER", "2"),
				mockBulkStrings("bar", "baz"),
			},
		},
		{
			name: "SCARD",
			ops: []TestOp{
				mockCmd("SCARD", "foo"),
				mockInt(0),
				mockCmd("SADD", "foo", "1"),
				mockSimpleString("OK"),
				mockCmd("SCARD", "foo"),
				mockInt(1),
				mockCmd("SADD", "foo", "2"),
				mockSimpleString("OK"),
				mockCmd("SCARD", "foo"),
				mockInt(2),
				mockCmd("SREM", "foo", "2"),
				mockInt(1),
				mockCmd("SCARD", "foo"),
				mockInt(1),
				mockCmd("DEL", "foo"),
				mockInt(1),
				mockCmd("SCARD", "foo"),
				mockInt(0),
			},
		},
		{
			name: "SINTERCARD",
			ops: []TestOp{
				mockCmd("SADD", "foo", "1"),
				mockSimpleString("OK"),
				mockCmd("SADD", "foo", "2"),
				mockSimpleString("OK"),
				mockCmd("SADD", "foo", "3"),
				mockSimpleString("OK"),
				mockCmd("SADD", "bar", "3"),
				mockSimpleString("OK"),
				mockCmd("SADD", "bar", "4"),
				mockSimpleString("OK"),
				mockCmd("SADD", "bar", "5"),
				mockSimpleString("OK"),
				mockCmd("SADD", "baz", "5"),
				mockSimpleString("OK"),
				mockCmd("SADD", "baz", "6"),
				mockSimpleString("OK"),
				mockCmd("SADD", "baz", "7"),
				mockSimpleString("OK"),
				mockCmd("SINTERCARD", "foo", "bar"),
				mockInt(1),
				mockCmd("SINTERCARD", "foo", "bar", "baz"),
				mockInt(0),
				mockCmd("SADD", "baz", "3"),
				mockSimpleString("OK"),
				mockCmd("SINTERCARD", "foo", "bar", "baz"),
				mockInt(1),
				mockCmd("SADD", "t", "1"),
				mockSimpleString("OK"),
				mockCmd("SADD", "t2", "2"),
				mockSimpleString("OK"),
				mockCmd("SINTERCARD", "t", "t2"),
				mockInt(0),
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
				t.Logf("running: %s", s2kv.CommandString(cmd))
				err := s2kv.CommandHandlers[string(cmd.Get(0))](db, writer, cmd)
				if err != nil {
					t.Error(err)
				}
			}
		})
	}

	t.Run("all commands have a test", func(t *testing.T) {
		expectedTestNames := make([]string, 0, len(s2kv.CommandHandlers))
		for name := range s2kv.CommandHandlers {
			expectedTestNames = append(expectedTestNames, name)
		}
		actualTestNames := make([]string, 0, len(tests))
		for _, test := range tests {
			actualTestNames = append(actualTestNames, test.name)
		}
		expect := gomega.ConsistOf(expectedTestNames)
		if matches, _ := expect.Match(actualTestNames); !matches {
			t.Error(expect.FailureMessage(actualTestNames))
		}
	})
}

func NewCmd(ctrl *gomock.Controller, args ...string) *MockCommand {
	cmd := NewMockCommand(ctrl)
	cmd.EXPECT().ArgCount().Return(len(args)).AnyTimes()
	for i, arg := range args {
		// we expect Get to be called at least twice
		// first to log the command in the test above
		// and second during the actual command processing
		cmd.EXPECT().Get(i).Return([]byte(arg)).MinTimes(2)
	}
	return cmd
}
