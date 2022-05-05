package s2kv

import (
	"strconv"
)

//go:generate mockgen -destination=mocks_test.go -package=s2kv_test . Command,Writer

type Command interface {
	Get(int) []byte
	ArgCount() int
}

type Writer interface {
	Write([]byte) (int, error)
	WriteBulk([]byte) error
	WriteBulks(...[]byte) error
	WriteBulkString(string) error
	WriteBulkStrings([]string) error
	WriteSimpleString(string) error
	WriteInt(int64) error
	WriteError(string) error
}

type CommandHandler func(*SingleStore, Writer, Command) error

var CommandHandlers = map[string]CommandHandler{
	"PING": func(_ *SingleStore, w Writer, c Command) error {
		return w.WriteSimpleString("PONG")
	},

	"SET": func(db *SingleStore, w Writer, c Command) error {
		key := string(c.Get(1))
		val := c.Get(2)

		err := db.BlobSet(key, val)
		if err != nil {
			return err
		}
		return w.WriteSimpleString("OK")
	},

	"INCRBY": func(db *SingleStore, w Writer, c Command) error {
		key := string(c.Get(1))
		val, err := strconv.ParseInt(string(c.Get(2)), 10, 64)
		if err != nil {
			return err
		}

		result, err := db.IncrBy(key, val)
		if err != nil {
			return err
		}
		return w.WriteInt(result)
	},

	"GET": func(db *SingleStore, w Writer, c Command) error {
		key := string(c.Get(1))
		val, err := db.BlobGet(key)
		if err != nil {
			return err
		}
		return w.WriteBulk(val)
	},

	"DEL": func(db *SingleStore, w Writer, c Command) error {
		key := string(c.Get(1))
		val, err := db.KeyDelete(key)
		if err != nil {
			return err
		}
		if val {
			return w.WriteInt(1)
		}
		return w.WriteInt(0)
	},

	"FLUSHALL": func(db *SingleStore, w Writer, c Command) error {
		err := db.FlushAll()
		if err != nil {
			return err
		}
		return w.WriteSimpleString("OK")
	},

	"KEYS": func(db *SingleStore, w Writer, c Command) error {
		pattern := string(c.Get(1))
		if pattern == "" {
			pattern = "%"
		}
		out, err := db.Keys(pattern)
		if err != nil {
			return err
		}
		return w.WriteBulks(out...)
	},

	"EXISTS": func(db *SingleStore, w Writer, c Command) error {
		key := string(c.Get(1))
		exists, err := db.KeyExists(key)
		if err != nil {
			return err
		}
		if exists {
			return w.WriteInt(1)
		}
		return w.WriteInt(0)
	},

	"RPUSH": func(db *SingleStore, w Writer, c Command) error {
		key := string(c.Get(1))
		val := c.Get(2)
		err := db.ListAppend(key, val)
		if err != nil {
			return err
		}
		return w.WriteSimpleString("OK")
	},

	"LREM": func(db *SingleStore, w Writer, c Command) error {
		key := string(c.Get(1))
		val := c.Get(2)
		n, err := db.ListRemove(key, val)
		if err != nil {
			return err
		}
		return w.WriteInt(n)
	},

	"LRANGE": func(db *SingleStore, w Writer, c Command) error {
		key := string(c.Get(1))
		start, err := strconv.Atoi(string(c.Get(2)))
		if err != nil {
			return err
		}
		stop, err := strconv.Atoi(string(c.Get(3)))
		if err != nil {
			return err
		}

		var out [][]byte
		if start == 0 && stop == -1 {
			out, err = db.ListGet(key)
			if err != nil {
				return err
			}
		} else {
			out, err = db.ListRange(key, start, stop)
			if err != nil {
				return err
			}
		}
		return w.WriteBulks(out...)
	},

	"SADD": func(db *SingleStore, w Writer, c Command) error {
		key := string(c.Get(1))
		val := c.Get(2)
		err := db.SetAdd(key, val)
		if err != nil {
			return err
		}
		return w.WriteSimpleString("OK")
	},

	"SREM": func(db *SingleStore, w Writer, c Command) error {
		key := string(c.Get(1))
		val := c.Get(2)
		n, err := db.SetRemove(key, val)
		if err != nil {
			return err
		}
		return w.WriteInt(n)
	},

	"SMEMBERS": func(db *SingleStore, w Writer, c Command) error {
		key := string(c.Get(1))
		out, err := db.SetGet(key)
		if err != nil {
			return err
		}
		return w.WriteBulks(out...)
	},

	"SUNION": func(db *SingleStore, w Writer, c Command) error {
		keys := commandSliceStr(c, 1, c.ArgCount())
		out, err := db.SetUnion(keys...)
		if err != nil {
			return err
		}
		return w.WriteBulks(out...)
	},

	"SINTER": func(db *SingleStore, w Writer, c Command) error {
		keys := commandSliceStr(c, 1, c.ArgCount())
		out, err := db.SetIntersect(keys...)
		if err != nil {
			return err
		}
		return w.WriteBulks(out...)
	},

	"SINTERCARD": func(db *SingleStore, w Writer, c Command) error {
		keys := commandSliceStr(c, 1, c.ArgCount())
		out, err := db.SetIntersectCardinality(keys...)
		if err != nil {
			return err
		}
		return w.WriteInt(out)
	},

	"SWITHMEMBER": func(db *SingleStore, w Writer, c Command) error {
		val := c.Get(1)
		out, err := db.SetsWithMember(val)
		if err != nil {
			return err
		}
		return w.WriteBulkStrings(out)
	},

	"SCARD": func(db *SingleStore, w Writer, c Command) error {
		key := string(c.Get(1))
		n, err := db.SetCardinality(key)
		if err != nil {
			return err
		}
		return w.WriteInt(n)
	},
}

func commandSliceStr(c Command, start, end int) []string {
	ret := make([]string, end-start)
	for i := start; i < end; i++ {
		ret[i-start] = string(c.Get(i))
	}
	return ret
}

func CommandString(c Command) string {
	ret := ""
	for i := 0; i < c.ArgCount(); i++ {
		if i > 0 {
			ret += " "
		}
		ret += string(c.Get(i))
	}
	return ret
}
