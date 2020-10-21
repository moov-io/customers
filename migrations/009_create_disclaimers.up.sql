create table disclaimers(
  disclaimer_id varchar(40) primary key, 
  text text, 
  document_id varchar(40), 
  created_at datetime, 
  deleted_at datetime
);
