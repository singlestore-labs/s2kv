drop database if exists kv;
create database kv;
use kv;

create rowstore table keyspace (
  k text,
  t enum("blob", "set", "list"),
  primary key (k)
);

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