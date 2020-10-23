create table validations(
  validation_id varchar(40) primary key, 
  account_id varchar(40), 
  status varchar(20), 
  strategy varchar(20), 
  vendor varchar(20), 
  created_at datetime(3), 
  updated_at datetime(3)
);
