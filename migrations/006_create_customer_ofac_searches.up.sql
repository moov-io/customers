create table if not exists customer_ofac_searches(
  customer_id varchar(40), 
  entity_id varchar(40), 
  sdn_name varchar(40), 
  sdn_type integer, 
  percentage_match double precision (5, 2), 
  blocked boolean,
  created_at datetime
);
