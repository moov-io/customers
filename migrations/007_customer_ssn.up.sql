create table customer_ssn(
  customer_id varchar(40) primary key not null,

  ssn         BLOB,
  ssn_masked  varchar(10) not null,

  created_at datetime not null
);
