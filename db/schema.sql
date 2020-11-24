create table blocks (
    hash char(66) primary key,
    number varchar not null,
    time bigint not null,
    parenthash char(66) not null,
    difficulty varchar not null,
    gasused bigint not null,
    gaslimit bigint not null,
    nonce bigint not null
);

create table transactions (
    hash char(66) primary key,
    from char(42) not null,
    to char(42) not null,
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
    data bytea not null,
    txhash char(66) not null,
    blockhash char(66) not null,
    primary key (blockhash, index),
    foreign key (txhash) references transactions(hash),
    foreign key (blockhash) references blocks(hash)
);

create table delivery_history (
    id bigserial primary key,
    client varchar(20) not null,
    ts timestamp not null,
    endpoint varchar(100) not null,
    datalength bigint not null
);

create table users (
    address char(42) not null,
    apikey char(66) primary key,
    ts timestamp not null
);

create table subscription_plans (
    id serial primary key,
    name varchar(20) not null,
    deliverycount bigint not null
);
