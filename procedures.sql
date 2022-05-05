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

create or replace procedure listAppend(_k text, _v text)
as begin
  start transaction;
  call assertKey(_k, "list");

  insert into listvalues (k, v) values (_k, _v);
  commit;
end //

create or replace procedure listRemove(_k text, _v text)
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

create or replace procedure setAdd(_k text, _v text)
as begin
  start transaction;
  call assertKey(_k, "set");
  insert ignore into setvalues (k, v) values (_k, _v);
  commit;
end //

create or replace procedure setRemove(_k text, _v text)
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

create or replace function setUnion2(_a text, _b text)
returns table as return select distinct(v) from setvalues where k in (_a, _b) //

create or replace function setUnion3(_a text, _b text, _c text)
returns table as return select distinct(v) from setvalues where k in (_a, _b, _c) //

create or replace function setUnion4(_a text, _b text, _c text, _d text)
returns table as return select distinct(v) from setvalues where k in (_a, _b, _c, _d) //

create or replace function setIntersect2(_a text, _b text)
returns table as return
    select distinct a.v
    from setvalues a, setvalues b
    where
      a.k = _a
      and b.k = _b and a.v = b.v //

create or replace function setIntersect3(_a text, _b text, _c text)
returns table as return
    select distinct a.v
    from setvalues a, setvalues b, setvalues c
    where
      a.k = _a
      and b.k = _b and a.v = b.v
      and c.k = _c and a.v = c.v //

create or replace function setIntersect4(_a text, _b text, _c text, _d text)
returns table as return
    select distinct a.v
    from setvalues a, setvalues b, setvalues c, setvalues d
    where
      a.k = _a
      and b.k = _b and a.v = b.v
      and c.k = _c and a.v = c.v
      and d.k = _d and a.v = d.v //

create or replace function setsWithMember(_v blob)
returns table as return
    select k from setvalues where v = _v //

create or replace function setCardinality(_k text)
returns table as return
    select count(*) from setvalues where k = _k //

create or replace function setIntersectCardinality2(_a text, _b text)
returns table as return
    select count(distinct(a.v))
    from setvalues a, setvalues b
    where
      a.k = _a
      and b.k = _b and a.v = b.v //

create or replace function setIntersectCardinality3(_a text, _b text, _c text)
returns table as return
    select count(*) from setIntersect3(_a, _b, _c) //

create or replace function setIntersectCardinality4(_a text, _b text, _c text, _d text)
returns table as return
    select count(*) from setIntersect4(_a, _b, _c, _d) //

delimiter ;