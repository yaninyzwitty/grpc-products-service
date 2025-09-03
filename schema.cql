-- Category table
CREATE TABLE IF NOT EXISTS categories (
  id bigint PRIMARY KEY,
  name text,
  description text,
  created_at timestamp
);


-- Product table
CREATE TABLE IF NOT EXISTS products (
  id bigint ,
  name text,
  description text,
  price float,
  stock int,
  category_id bigint,
  created_at timestamp,
  updated_at timestamp,
  PRIMARY KEY ((category_id), id)
) WITH CLUSTERING ORDER BY (id DESC);




CREATE TABLE IF NOT EXISTS products_outbox (
    id UUID,
    bucket TEXT,
    payload TEXT,
    event_type TEXT,
    PRIMARY KEY((bucket), id)
);

