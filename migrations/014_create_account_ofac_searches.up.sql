create table if not exists account_ofac_searches(
  account_ofac_search_id varchar(40) primary key, 
  account_id varchar(40), 
  entity_id varchar(40), 
  sdn_name varchar(40), 
  sdn_type integer, 
  percentage_match double precision (5, 2), 
  created_at datetime
);
