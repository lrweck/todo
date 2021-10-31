CREATE TYPE priority AS ENUM ('none', 'low', 'medium', 'high');

CREATE TABLE tasks (
  id          UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  description VARCHAR NOT NULL,
  priority    priority DEFAULT 'none' NOT NULL,
  start_date  TIMESTAMP WITHOUT TIME ZONE,
  due_date    TIMESTAMP WITHOUT TIME ZONE,
  done        BOOLEAN NOT NULL DEFAULT FALSE
);