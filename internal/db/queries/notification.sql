-- db/queries/notification.sql
-- name: CreateNotification :one
INSERT INTO
    notifications (
        user_id,
        type,
        title,
        message
    )
VALUES
    ($1, $2, $3, $4)
RETURNING
    *;

-- name: MarkNotificationAsRead :one
UPDATE notifications
SET
    is_read = TRUE,
    read_at = now()
WHERE
    id = $1
    AND user_id = $2
RETURNING
    *;

-- name: GetUserNotifications :many
SELECT
    *
FROM
    notifications
WHERE
    user_id = $1
ORDER BY
    created_at DESC;