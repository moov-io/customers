create table customer_status_updates(
  customer_id varchar(40) not null,

  future_status integer not null,
  comment varchar(512),
  changed_at datetime not null

  -- TODO(adam): create index on customer_id
);
