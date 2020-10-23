create table customer_status_updates(
  customer_id varchar(40), 
  future_status integer, 
  comment varchar(512), 
  changed_at datetime(6)
);
