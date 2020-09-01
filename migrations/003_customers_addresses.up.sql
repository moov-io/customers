create table customers_addresses(
  address_id  varchar(40) primary key not null,
  customer_id varchar(40) not null,

  type        SMALLINT not null,
  address1    varchar(120) not null,
  address2    varchar(120) not null,
  city        varchar(50) not null,
  state       varchar(2) not null,
  postal_code varchar(10) not null,
  country     varchar(3) not null,
  validated   boolean default false,

  constraint customer_address1_uniq unique (customer_id, address1) on conflict abort
);
