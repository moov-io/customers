create table customer_ofac_searches(
  customer_id varchar(40) not null,

  entity_id   varchar(40) not null,
  sdn_name    varchar(512) not null,
  sdn_type    integer not null,
  match       double precision (5,2) not null,

  created_at datetime not null
);
