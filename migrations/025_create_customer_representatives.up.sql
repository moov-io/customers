CREATE TABLE if not exists customer_representatives (
    representative_id varchar(40) primary key,
    customer_id varchar(40),
    first_name varchar(40),
    last_name varchar(40),
    job_title varchar(50),
    birth_date datetime,
    created_at datetime,
    deleted_at datetime
);
