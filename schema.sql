drop database if exists kv;
create database kv;
use kv;

create table blobvalues (
  k text,
  v blob,

  shard (k),
  sort key (),
  unique key (k) using hash
);

create table setvalues (
  k text,
  v blob,

  shard (k),
  sort key (v),
  unique key (k, v) using hash
);

create table listvalues (
  k text not null,
  v blob not null,

  -- ordering columns
  ts datetime(6) not null default now(),
  seq bigint not null auto_increment,

  shard (k),
  sort key (k, ts, seq),
  unique key (k, ts, seq) using hash
);

create view keyspace as (
  select k from blobvalues
  union
  select k from setvalues
  union
  select k from listvalues
);
