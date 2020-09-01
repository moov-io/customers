create table customer_metadata(
  customer_id varchar(40) primary key not null,

  meta_key      varchar(40) not null,
  meta_value    varchar(512) not null,

  constraint customer_metadata_uniq unique (customer_id, meta_key, meta_value)
);
