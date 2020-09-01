create table documents(
  document_id varchar(40) primary key not null,
  customer_id varchar(40) not null,

  type         varchar(120) not null,
  content_type integer not null,

  uploaded_at datetime not null
);
