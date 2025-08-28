CREATE TABLE IF NOT EXISTS roles (
    id bigserial PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    level int NOT NULL DEFAULT 0,
    description text
);

INSERT INTO
    roles (name,description,level)
    VALUES (
        'user',
        'A user create posts and comments',
        1
    );

INSERT INTO
    roles (name,description,level)
    VALUES (
        'moderator',
        'A moderator can update other users posts',
        2
    );


INSERT INTO
    roles (name,description,level)
    VALUES (
        'admin',
        'A admin can update other users posts',
        3
    );