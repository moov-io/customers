create table customers(
  customer_id varchar(40) primary key not null,

  first_name      varchar(255) not null,
  middle_name     varchar(255),
  last_name       varchar(255) not null,
  nick_name       varchar(255),
  suffix          varchar(20),
  birth_date      timestamp,
  status          varchar(20) not null,
  email           varchar(255) not null,
  type            varchar(20) not null,

  created_at     timestamp not null,
  last_modified  timestamp not null,
  deleted_at     timestamp,

  constraint email_uniq unique (email)
);
