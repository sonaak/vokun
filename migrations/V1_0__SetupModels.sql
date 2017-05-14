-- creating users
CREATE USER vokun_reader WITH NOSUPERUSER ENCRYPTED PASSWORD '${reader_password}';
CREATE USER vokun_writer WITH NOSUPERUSER ENCRYPTED PASSWORD '${writer_password}';

-- Request: objects representing a typical request
-- that isn't a streaming request
CREATE TABLE models.request (
  id UUID PRIMARY KEY,
  subpath VARCHAR,
  request_time TIMESTAMP WITHOUT TIME ZONE,
  body BYTEA
);

-- Headers: objects representing key/val sent with
-- a request in the HTML header
CREATE TABLE models.header (
  id UUID PRIMARY KEY,
  request_id UUID REFERENCES request (id),
  name VARCHAR,
  value VARCHAR
);

-- Headers: objects representing key/val sent with
-- a request in the HTML header
CREATE TABLE models.params (
  id UUID PRIMARY KEY,
  request_id UUID REFERENCES request (id),
  name VARCHAR,
  value VARCHAR
);

-- apply grants to the schema and tables
GRANT USAGE ON SCHEMA models TO vokun_reader, vokun_writer;
GRANT SELECT ON ALL TABLES IN SCHEMA models TO vokun_reader;
GRANT SELECT,INSERT,UPDATE,DELETE,TRIGGER ON ALL TABLES IN SCHEMA models TO vokun_writer;
