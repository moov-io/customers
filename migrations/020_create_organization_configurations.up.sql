create table if not exists organization_configuration(
  organization varchar(40) primary key not null, 
  legal_entity varchar(40) not null, 
  primary_account varchar(40) not null
);
