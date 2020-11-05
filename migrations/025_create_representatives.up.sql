CREATE TABLE if not exists representatives (
    representative_id varchar(40) primary key,
    customer_id varchar(40),
    first_name varchar(40),
    last_name varchar(40),
    job_title varchar(50),
    birth_date datetime,
    created_at datetime,
    last_modified datetime,
    deleted_at datetime
);
