create table documents(
  document_id varchar(40) primary key, 
  customer_id varchar(40), 
  type varchar(120), 
  content_type varchar(40), 
  uploaded_at datetime(6),
  deleted_at datetime(6)
);
