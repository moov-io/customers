create table if not exists validations(
  validation_id varchar(40) primary key, 
  account_id varchar(40), 
  status varchar(20), 
  strategy varchar(20), 
  vendor varchar(20), 
  created_at datetime, 
  updated_at datetime
);

create index idx_validations_validation_account_ids on validations (validation_id, account_id)
