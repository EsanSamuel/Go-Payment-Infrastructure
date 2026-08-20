package jobs

import (
	"context"
	"fmt"
	"log"

	"example.com/internal/models"
	"github.com/gocraft/work"
	"github.com/jackc/pgx/v5/pgtype"
)

type NotificationInterface interface {
	CreateNotification(ctx context.Context, req models.CreateNotificationRequest) (*models.NotificationResponse, error)
}

type NotifContext struct {
	UserID              string
	NotifType           string
	Title               string
	Message             string
	TransactionID       string
	notificationService NotificationInterface
}

var (
	enqueuer *work.Enqueuer = work.NewEnqueuer("notification", RedisPool)
	notifSvc NotificationInterface
)

func EnqueueNotification(userID, notifType, title, message, transactionID string) {
	_, err := enqueuer.Enqueue("send_notification", work.Q{
		"user_id":        userID,
		"type":           notifType,
		"title":          title,
		"message":        message,
		"transaction_id": transactionID,
	})
	if err != nil {
		log.Printf("failed to enqueue notification for transaction %s: %v", transactionID, err)
	}
}

func NotificationWorker(svc NotificationInterface) {
	notifSvc = svc

	worker := work.NewWorkerPool(NotifContext{}, 10, "notification", RedisPool)
	worker.Middleware((*NotifContext).Log)
	worker.Middleware((*NotifContext).InjectNotificationService)
	worker.Middleware((*NotifContext).FindNotificationPayload)
	worker.Job("send_notification", (*NotifContext).SendNotification)
	worker.Start()
}

func StopNotificationWorker() {
	worker := work.NewWorkerPool(NotifContext{}, 10, "notification", RedisPool)
	worker.Stop()
}

func (c *NotifContext) Log(job *work.Job, next work.NextMiddlewareFunc) error {
	fmt.Println("Background job is running:", job.Name, "ID:", job.ID)
	return next()
}

// WHAT I LEARNT
// Inject supplies the service dependency — gocraft/work rebuilds Context
// from scratch per job via reflection, so field values set at construction
// time never survive. This is the only way to get it in.
func (c *NotifContext) InjectNotificationService(job *work.Job, next work.NextMiddlewareFunc) error {
	c.notificationService = notifSvc
	return next()
}

func (c *NotifContext) FindNotificationPayload(job *work.Job, next work.NextMiddlewareFunc) error {
	c.UserID = job.ArgString("user_id")
	c.NotifType = job.ArgString("type")
	c.Title = job.ArgString("title")
	c.Message = job.ArgString("message")
	c.TransactionID = job.ArgString("transaction_id")
	if err := job.ArgError(); err != nil {
		return err
	}
	return next()
}

func (c *NotifContext) SendNotification(job *work.Job) error {
	var uid pgtype.UUID
	if err := uid.Scan(c.UserID); err != nil {
		return fmt.Errorf("invalid user_id %q in job %s: %w", c.UserID, c.TransactionID, err)
	}

	_, err := c.notificationService.CreateNotification(context.Background(), models.CreateNotificationRequest{
		UserID:  uid,
		Type:    c.NotifType,
		Title:   c.Title,
		Message: c.Message,
	})
	if err != nil {
		return fmt.Errorf("notification failed for transaction %s: %w", c.TransactionID, err)
	}
	return nil
}
