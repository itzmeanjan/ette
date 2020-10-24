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
