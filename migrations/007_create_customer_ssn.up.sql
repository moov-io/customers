create table if not exists customer_ssn(
  customer_id varchar(40) primary key, 
  ssn BLOB, 
  ssn_masked varchar(9), 
  created_at datetime
);
