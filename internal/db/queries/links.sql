-- db/queries/links.sql

-- name: GetLink :one
SELECT * FROM links WHERE id = $1;

-- name: GetLinks :many
SELECT * FROM links ORDER BY created_at;

-- name: CreateLink :one
INSERT INTO links (title, url, user_id) VALUES ($1, $2, $3) RETURNING *;

-- name: ListLinksWithTags :many
SELECT l.*, ARRAY_AGG(t.name) AS tags
FROM links l
LEFT JOIN link_tags lt ON lt.link_id = l.id
LEFT JOIN tags t ON t.id = lt.tag_id
WHERE l.user_id = $1
GROUP BY l.id;

-- name: GetLinksWithUserDetails :many
SELECT l.id,
    l.title,
    l.url,
    l.created_at,
    u.id AS user_id,
    u.email AS user_email
FROM links l
INNER JOIN users u ON l.user_id=u.id;

-- name: GetMostVisitedLinks :many
SELECT l.*, COUNT(v.id) AS visit_count
FROM links l
JOIN link_visits v ON v.link_id = l.id
GROUP BY l.id
ORDER BY visit_count DESC
LIMIT 10;
