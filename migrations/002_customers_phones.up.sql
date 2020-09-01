create table customers_phones(
  customer_id varchar(40) primary key not null,

  number  varchar(20) not null,
  valid   boolean default false,
  type    varchar(20) not null,

  constraint customer_number_uniq unique (customer_id, number)
);
