-- +migrate Up
CREATE TABLE users (
 uuid UUID PRIMARY KEY,
 created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
 updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
 deleted_at TIMESTAMP WITH TIME ZONE,
 account_id UUID REFERENCES accounts (uuid),
 first_name TEXT,
 last_name TEXT,
 email TEXT NOT NULL,
 auth0_id VARCHAR(100) NOT NULL
);

-- +migrate Down
DROP TABLE users;
