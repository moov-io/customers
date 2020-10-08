create table if not exists customer_metadata(
  customer_id varchar(40), 
  meta_key varchar(40), 
  meta_value varchar(512), 
  unique(meta_key, meta_value)
);
