create table disclaimers(
  disclaimer_id varchar(40) primary key not null,

  text         text not null,
  document_id varchar(40),

  created_at datetime not null,
  deleted_at datetime
);
