use kv;

delimiter //

create or replace procedure assertKeyIsList (_k text)
as
declare
  _q QUERY(b boolean) = select exists(select 1 from blobvalues where k = _k) or exists (select 1 from setvalues where k = _k);
begin
  if scalar(_q) then
    RAISE user_exception("key exists but is not a list");
  end if;
end //

create or replace procedure assertKeyIsBlob (_k text)
as
declare
  _q QUERY(b boolean) = select exists(select 1 from listvalues where k = _k) or exists (select 1 from setvalues where k = _k);
begin
  if scalar(_q) then
    RAISE user_exception("key exists but is not a blob");
  end if;
end //

create or replace procedure assertKeyIsSet (_k text)
as
declare
  _q QUERY(b boolean) = select exists(select 1 from listvalues where k = _k) or exists (select 1 from blobvalues where k = _k);
begin
  if scalar(_q) then
    RAISE user_exception("key exists but is not a set");
  end if;
end //

create or replace procedure keyExists (_k text)
returns boolean as
declare
  _q QUERY(b boolean) = select exists(select * from keyspace where k = _k);
begin
  return scalar(_q);
end //

create or replace procedure keyDelete (_k text)
returns boolean as
begin
  start transaction;

  delete from blobvalues where k = _k;
  if row_count() > 0 then
    commit;
    return true;
  end if;

  delete from listvalues where k = _k;
  if row_count() > 0 then
    commit;
    return true;
  end if;

  delete from listvalues where k = _k;
  if row_count() > 0 then
    commit;
    return true;
  end if;

  rollback;
  return false;
end //

create or replace procedure blobSet (_k text, _v blob)
as
begin
  start transaction;
  call assertKeyIsBlob(_k);

  insert into blobvalues (k, v) values (_k, _v)
    on duplicate key update v = values(v);
  commit;

exception when others then rollback; raise;
end //

create or replace procedure blobGet (_k text)
returns query(b blob) as
declare
  _q query(b blob) = select v from blobvalues where k = _k;
begin
  call assertKeyIsBlob(_k);
  return _q;
end //

create or replace procedure listAppend(_k text, _v text)
as
begin
  start transaction;
  call assertKeyIsList(_k);

  insert into listvalues (k, v) values (_k, _v);
  commit;

exception when others then rollback; raise;
end //

create or replace procedure listRemove(_k text, _v text)
returns int as
declare
  _rowcount int;
begin
  start transaction;
  call assertKeyIsList(_k);

  delete from listvalues where k = _k and v = _v;
  _rowcount = row_count();
  commit;

  return _rowcount;

exception when others then rollback; raise;
end //

create or replace procedure listGet(_k text)
returns query(v text) as
declare
  _q query(v text) = select * from listvalues where k = _k;
begin
  call assertKeyIsSet(_k);
  return _q;
end //

create or replace procedure listRange(_k text, _start int, _end int)
returns QUERY(v text) as
declare
  _return QUERY(v text) =
    select v
    from (
      select v, row_number() over (order by ts, seq) as _rownum
      from listvalues
      where k = _k
    )
    where _rownum >= _start and _rownum <= _end
    order by _rownum asc;
begin
  call assertKeyIsList(_k);
  return _return;
end //

create or replace procedure setAdd(_k text, _v text)
as
begin
  start transaction;
  call assertKeyIsSet(_k);

  insert ignore into setvalues (k, v) values (_k, _v);
  commit;

exception when others then rollback; raise;
end //

create or replace procedure setRemove(_k text, _v text)
as
begin
  start transaction;
  call assertKeyIsSet(_k);

  delete from setvalues where k = _k and v = _v;
  commit;

exception when others then rollback; raise;
end //

create or replace procedure setGet(_k text)
returns query(v text) as
declare
  _q query(v text) = select v from setvalues where k = _k;
begin
  call assertKeyIsSet(_k);
  return _q;
end //

create or replace procedure setUnion2(_a text, _b text)
returns query(v text) as
declare
  _q query(v text) = select distinct(v) from setvalues where k in (_a, _b);
begin
  -- TODO: need to assert all _keys are SETs
  return _q;
end //

create or replace procedure setUnion3(_a text, _b text, _c text)
returns query(v text) as
declare
  _q query(v text) = select distinct(v) from setvalues where k in (_a, _b, _c);
begin
  -- TODO: need to assert all _keys are SETs
  return _q;
end //

create or replace procedure setUnion4(_a text, _b text, _c text, _d text)
returns query(v text) as
declare
  _q query(v text) = select distinct(v) from setvalues where k in (_a, _b, _c, _d);
begin
  -- TODO: need to assert all _keys are SETs
  return _q;
end //

create or replace procedure setIntersect2(_a text, _b text)
returns query(v text) as
declare
  _q query(v text) =
    select distinct a.v
    from setvalues a, setvalues b
    where
      a.k = _a
      and b.k = _b and a.v = b.v;
begin
  -- TODO: need to assert all _keys are SETs
  return _q;
end //

create or replace procedure setIntersect3(_a text, _b text, _c text)
returns query(v text) as
declare
  _q query(v text) =
    select distinct a.v
    from setvalues a, setvalues b, setvalues c
    where
      a.k = _a
      and b.k = _b and a.v = b.v
      and c.k = _c and a.v = c.v;
begin
  -- TODO: need to assert all _keys are SETs
  return _q;
end //

create or replace procedure setIntersect4(_a text, _b text, _c text, _d text)
returns query(v text) as
declare
  _q query(v text) =
    select distinct a.v
    from setvalues a, setvalues b, setvalues c, setvalues d
    where
      a.k = _a
      and b.k = _b and a.v = b.v
      and c.k = _c and a.v = c.v
      and d.k = _d and a.v = d.v;
begin
  -- TODO: need to assert all _keys are SETs
  return _q;
end //

delimiter ;

delete from blobvalues;
delete from listvalues;
delete from setvalues;

select "appending values";
call listAppend(1,1);
call listAppend(1,2);
call listAppend(1,3);
call listAppend(1,3);
select "checking list";
select k, v from listvalues where k = 1 order by ts, seq;

select "list range 0,10";
echo listRange(1,0,10);
select "list range 1,10";
echo listRange(1,1,10);
select "list range 1,2";
echo listRange(1,1,2);
select "list range 2,4";
echo listRange(1,2,4);

select "removing (1,3)";
echo listRemove(1,3);
select k, v from listvalues where k = 1 order by ts, seq;

select "removing (1,1)";
echo listRemove(1,1);
select k, v from listvalues where k = 1 order by ts, seq;

select "removing (1,1) (again)";
echo listRemove(1,1);
select k, v from listvalues where k = 1 order by ts, seq;

select "keyExists(1)";
echo keyExists(1);
select "keyDelete(1)";
echo keyDelete(1);
select "keyExists(1)";
echo keyExists(1);

select "blob functions";
call blobSet(1,1);
echo blobGet(1);
call blobSet(2,2);
echo blobGet(2);
call blobSet(2,3);
echo blobGet(2);

-- should fail
call listAppend(1,1);

-- should succeed
call listAppend(3,1);

-- should fail
call blobSet(3,1);

delete from blobvalues;
delete from listvalues;
delete from setvalues;

call setAdd(1,1);
call setAdd(1,1);
call setAdd(1,2);
call setAdd(1,3);

call setAdd(2,3);
call setAdd(2,4);
call setAdd(2,5);

echo setGet(1);
echo setUnion2(1,2); -- should be 1,2,3,4,5
echo setIntersect2(1,2); -- should be 3
