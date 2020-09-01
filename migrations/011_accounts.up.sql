create table accounts(
  account_id   varchar(40) primary key not null,
  customer_id  varchar(40) not null,
  user_id      varchar(40) not null, -- uhh...

  encrypted_account_number varchar(100) not null,
  hashed_account_number    varchar(50) not null,
  masked_account_number    varchar(20) not null,
  routing_number           varchar(10) not null,
  status                   varchar(12) not null,
  type                     varchar(12) not null,
  holder_name              varchar(60) not null,

  created_at datetime not null,
  deleted_at datetime,

  constraint unique_accounts_to_customer UNIQUE (customer_id, hashed_account_number, routing_number)
);
