# In procedures.sql

```sql
create or replace procedure decrBy (_k text, _v bigint) returns bigint
as begin
  return incrBy(_k, -1 * _v);
end //
```

# In database.go

```go
func (s *SingleStore) DecrBy(k string, v int64) (int64, error) {
	var out int64
	err := s.db.Get(&out, "echo decrBy(?, ?)", k, v)
	if err != nil {
		return 0, err
	}
	return out, nil
}
```

# In commands.go

```go
	"DECRBY": func(db *SingleStore, w Writer, c Command) error {
		key := string(c.Get(1))
		val, err := strconv.ParseInt(string(c.Get(2)), 10, 64)
		if err != nil {
			return err
		}

		result, err := db.DecrBy(key, val)
		if err != nil {
			return err
		}
		return w.WriteInt(result)
	},
```

# In commands_test.go

```go
		{
			name: "DECRBY",
			ops: []TestOp{
				mockCmd("DECRBY", "foo", "0"),
				mockInt(0),
				mockCmd("GET", "foo"),
				mockBulk("0"),
				mockCmd("DECRBY", "foo", "1"),
				mockInt(-1),
				mockCmd("GET", "foo"),
				mockBulk("-1"),
				mockCmd("DECRBY", "foo", "10"),
				mockInt(-11),
				mockCmd("GET", "foo"),
				mockBulk("-11"),
				mockCmd("DECRBY", "foo", "-5"),
				mockInt(-6),
				mockCmd("GET", "foo"),
				mockBulk("-6"),
				mockCmd("DECRBY", "bar", "-5"),
				mockInt(5),
				mockCmd("GET", "bar"),
				mockBulk("5"),
				mockCmd("DECRBY", "baz", "100"),
				mockInt(-100),
				mockCmd("GET", "baz"),
				mockBulk("-100"),
			},
		},
```