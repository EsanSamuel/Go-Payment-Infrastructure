-- db/migrations/00004_create_link_tags.up.sql
CREATE TABLE
    link_tags (
        link_id BIGINT NOT NULL REFERENCES links (id) ON DELETE CASCADE,
        tag_id BIGINT NOT NULL REFERENCES tags (id) ON DELETE CASCADE,
        PRIMARY KEY (link_id, tag_id)
    )