create table if not exists disclaimer_acceptances(
  disclaimer_id varchar(40), 
  customer_id varchar(40), 
  accepted_at datetime, 
  unique(disclaimer_id, customer_id)
);
