use kv;

delimiter //

create or replace function keyExists (_k text)
returns table as return
  select exists(select 1 from keyspace where k = _k) //

create or replace function getKeys (_pattern text)
returns table as return
  select k from keyspace where k like _pattern //

create or replace procedure keyDelete (_k text)
returns boolean as
begin
  start transaction;

  delete from blobvalues where k = _k;
  delete from listvalues where k = _k;
  delete from setvalues where k = _k;

  delete from keyspace where k = _k;
  if row_count() > 0 then
    commit;
    return true;
  end if;

  rollback;
  return false;
end //

create or replace procedure flushAll ()
as begin
  start transaction;
  delete from keyspace;
  delete from blobvalues;
  delete from listvalues;
  delete from setvalues;
  commit;
end //

create or replace function assertType (
  _actual enum("blob", "set", "list"),
  _expected enum("blob", "set", "list")
) returns text
as begin
  if _actual != _expected then
    raise user_exception(concat("type mismatch; got ", _actual, ", expected ", _expected));
  end if;
  return _actual;
end //

-- assertKey must be used within a transaction
-- will rollback the parent transaction on failure
create or replace procedure assertKey (_k text, _type enum("blob", "set", "list"))
as
declare
  _q query(t text) = select (select t from keyspace where k = _k);
  _actual_type text;
begin
  _actual_type = scalar(_q);

  if _actual_type is null then
    -- new key
    insert into keyspace (k, t) values (_k, _type)
      on duplicate key update t = assertType(t, _type);
  elsif _actual_type != _type then
    raise user_exception(concat("type mismatch; got ", _actual_type, ", expected ", _type));
  end if;

exception when others then rollback; raise;
end //

create or replace procedure blobSet (_k text, _v blob)
as begin
  start transaction;
  call assertKey(_k, "blob");

  insert into blobvalues (k, v) values (_k, _v)
    on duplicate key update v = values(v);

  commit;
end //

create or replace function blobGet (_k text)
returns table as return
  select (select v from blobvalues where k = _k) as v //

create or replace function assertNotNull (_v blob) returns blob
as begin
  if _v is null then
    raise user_exception("invalid operation");
  end if;
  return _v;
end //

create or replace procedure incrBy (_k text, _v bigint) returns bigint
as
declare
  _ret_q query(v bigint) = select v :> bigint from blobvalues where k = _k;
  _ret bigint;
begin
  start transaction;
  call assertKey(_k, "blob");

  insert into blobvalues (k, v) values (_k, _v)
    on duplicate key update v = assertNotNull(v + _v);

  _ret = scalar(_ret_q);

  commit;

  return _ret;
end //

create or replace procedure listAppend(_k text, _v blob)
as begin
  start transaction;
  call assertKey(_k, "list");

  insert into listvalues (k, v) values (_k, _v);
  commit;
end //

create or replace procedure listRemove(_k text, _v blob)
returns int as
declare
  _rowcount int;
begin
  start transaction;
  call assertKey(_k, "list");

  delete from listvalues where k = _k and v = _v;
  _rowcount = row_count();
  commit;

  return _rowcount;
end //

create or replace function listGet(_k text)
returns table as return
  select v from listvalues
    where k = _k
    order by ts, seq
  //

-- retrieves elements of list between offset _start and _end (inclusive)
-- _start is 0 based
create or replace function listRange(_k text, _start int, _end int)
returns table as return
  select v
  from (
    select v, (row_number() over (order by ts, seq)) - 1 as _rownum
    from listvalues
    where k = _k
  )
  where _rownum >= _start and _rownum <= _end
  order by _rownum asc //

create or replace procedure setAdd(_k text, _v blob)
as begin
  start transaction;
  call assertKey(_k, "set");
  insert ignore into setvalues (k, v) values (_k, _v);
  commit;
end //

create or replace procedure setRemove(_k text, _v blob)
returns int as
declare
  _rowcount int;
begin
  start transaction;
  call assertKey(_k, "set");
  delete from setvalues where k = _k and v = _v;
  _rowcount = row_count();
  commit;
  return _rowcount;
end //

create or replace function setGet(_k text)
returns table as return
  select v from setvalues where k = _k //

create or replace procedure setUnion(_keys array(text))
returns query(v blob) as
declare
  _q text = "select distinct(v) from setvalues where k in (";
begin
  if length(_keys) < 2 then
    raise user_exception("setUnion requires at least 2 keys");
  end if;

  for i in 0 .. length(_keys) - 1 loop
    if i > 0 then
      _q = concat(_q, ",");
    end if;

    _q = concat(_q, quote(_keys[i]));
  end loop;

  _q = concat(_q, ")");

  return to_query(_q);
end //

create or replace procedure setIntersect(_keys array(text))
returns query(v blob) as
declare
  _prefix text = "select distinct(s0.v) from ";
  _tables text = " ";
  _joins text = " ";
begin
  if length(_keys) < 2 then
    raise user_exception("setIntersect requires at least 2 keys");
  end if;

  for i in 0 .. length(_keys) - 1 loop
    if i = 0 then
      _tables = concat(_tables, "setvalues s0");
      _joins = concat(_joins, "s0.k = ", quote(_keys[0]));
    else
      _tables = concat(_tables, ", setvalues s", i);
      _joins = concat(
        _joins,
        -- and s1.k = _keys[1]
        " and s", i, ".k = ", quote(_keys[i]),
        -- and s0.v = s1.v
        " and s0.v = s", i, ".v"
      );
    end if;
  end loop;

  return to_query(concat(_prefix, _tables, " where ", _joins));
end //

create or replace function setsWithMember(_v blob)
returns table as return
    select k from setvalues where v = _v //

create or replace function setCardinality(_k text)
returns table as return
    select count(*) from setvalues where k = _k //

create or replace procedure setIntersectCardinality(_keys array(text))
returns query(c bigint) as
declare
  _base query(v blob) = setIntersect(_keys);
  _q query(c bigint) = select count(*) from _base;
begin return _q;
end //

delimiter ;