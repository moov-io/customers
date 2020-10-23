create table customers_phones (
  customer_id VARCHAR(40), 
  number VARCHAR(20), 
  valid BOOLEAN, 
  type integer,
  constraint customer_number unique (customer_id, number)
);
