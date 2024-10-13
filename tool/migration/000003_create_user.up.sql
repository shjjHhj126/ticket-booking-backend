-- Create the users table
CREATE TABLE IF NOT EXISTS users(
    id bigserial PRIMARY KEY,
    username text NOT NULL,
    password_hash text NOT NULL,
    email text NOT NULL UNIQUE, 
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);

-- Create the sections table
CREATE TABLE IF NOT EXISTS sessions (
    id bigserial PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_token text NOT NULL UNIQUE,
    created_at timestamp with time zone DEFAULT now(),
    expires_at timestamp with time zone NOT NULL
);
