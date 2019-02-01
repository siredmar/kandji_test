-- +migrate Up
CREATE TABLE accounts (
 uuid UUID PRIMARY KEY NOT NULL,
 created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
 updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
 deleted_at TIMESTAMP WITH TIME ZONE,
 account_name VARCHAR(500)
);

-- +migrate Down
DROP TABLE accounts;
