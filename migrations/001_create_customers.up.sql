create table customers (
  customer_id varchar(40), 
  first_name varchar(40), 
  middle_name varchar(40), 
  last_name varchar(40), 
  nick_name varchar(40), 
  suffix varchar(3), 
  birth_date datetime, 
  status varchar(20), 
  email varchar(120), 
  type varchar(25),
  organization varchar(40) not null,
  created_at datetime(6), 
  last_modified datetime(6), 
  deleted_at datetime(6), 
  primary key (customer_id)
);
