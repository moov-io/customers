create table disclaimer_acceptances(
  disclaimer_id varchar(40) primary key not null,
  customer_id   varchar(40) not null,

  accepted_at datetime not null,

  constraint customer_disclaimer unique (disclaimer_id, customer_id)
);
