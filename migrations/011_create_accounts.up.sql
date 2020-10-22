create table accounts(
  account_id varchar(40) primary key, 
  customer_id varchar(40), 
  user_id varchar(40), 
  encrypted_account_number varchar(100), 
  hashed_account_number varchar(64), 
  masked_account_number varchar(15), 
  routing_number varchar(10), 
  holder_name varchar(60) default '',
  status varchar(12), 
  type varchar(12), 
  created_at datetime, 
  deleted_at datetime,
  constraint accounts_unique_to_customer unique (customer_id, hashed_account_number, routing_number)
);

