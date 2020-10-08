create table if not exists customers_phones (
  customer_id VARCHAR(40), 
  number VARCHAR(20), 
  valid BOOLEAN, 
  type integer, 
  unique (customer_id, number)
);
