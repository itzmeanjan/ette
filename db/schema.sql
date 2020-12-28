-- Note : This file is never going to be used by `ette`
--
-- This is provided more sake of better human readability

create table blocks (
    hash char(66) primary key,
    number varchar not null unique,
    time bigint not null,
    parenthash char(66) not null,
    difficulty varchar not null,
    gasused bigint not null,
    gaslimit bigint not null,
    nonce bigint not null,
    miner char(42) not null,
    size float(8) not null,
    txroothash char(66) not null,
    receiptroothash char(66) not null,
);

create index on blocks(number asc);
create index on blocks(time asc);

create table transactions (
    hash char(66) primary key,
    from char(42) not null,
    to char(42),
    contract char(42),
    value varchar,
    data bytea,
    gas bigint not null,
    gasprice varchar not null,
    cost varchar not null,
    nonce bigint not null,
    state smallint not null,
    blockhash char(66) not null,
    foreign key (blockhash) references blocks(hash)
);

create table events (
    origin char(42) not null,
    index integer not null,
    topics text[] not null,
    data bytea,
    txhash char(66) not null,
    blockhash char(66) not null,
    primary key (blockhash, index),
    foreign key (txhash) references transactions(hash),
    foreign key (blockhash) references blocks(hash)
);

create table users (
    address char(42) not null,
    apikey char(66) primary key,
    ts timestamp not null,
    enabled boolean default true
);

create index on users(address);

create table delivery_history (
    id bigserial primary key,
    client char(42) not null,
    ts timestamp not null,
    endpoint varchar(100) not null,
    datalength bigint not null,
);

create index on delivery_history(client);

create table subscription_plans (
    id serial primary key,
    name varchar(20) not null unique,
    deliverycount bigint not null unique
);

create table subscription_details (
    address char(42) primary key,
    subscriptionplan int not null
);
