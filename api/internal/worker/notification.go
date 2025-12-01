package worker

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/tkasuz/s3local/internal/db"
)

// S3 Event Message Structures
type S3EventNotification struct {
	Records []S3EventRecord `json:"Records"`
}

type S3EventRecord struct {
	EventVersion      string                       `json:"eventVersion"`
	EventSource       string                       `json:"eventSource"`
	AWSRegion         string                       `json:"awsRegion"`
	EventTime         string                       `json:"eventTime"`
	EventName         string                       `json:"eventName"`
	UserIdentity      S3UserIdentity               `json:"userIdentity"`
	RequestParameters S3RequestParameters          `json:"requestParameters"`
	ResponseElements  S3ResponseElements           `json:"responseElements"`
	S3                S3Entity                     `json:"s3"`
	GlacierEventData  *S3GlacierEventData          `json:"glacierEventData,omitempty"`
}

type S3UserIdentity struct {
	PrincipalID string `json:"principalId"`
}

type S3RequestParameters struct {
	SourceIPAddress string `json:"sourceIPAddress"`
}

type S3ResponseElements struct {
	XAmzRequestID string `json:"x-amz-request-id"`
	XAmzID2       string `json:"x-amz-id-2"`
}

type S3Entity struct {
	S3SchemaVersion string   `json:"s3SchemaVersion"`
	ConfigurationID string   `json:"configurationId"`
	Bucket          S3Bucket `json:"bucket"`
	Object          S3Object `json:"object"`
}

type S3Bucket struct {
	Name          string        `json:"name"`
	OwnerIdentity S3OwnerIdentity `json:"ownerIdentity"`
	ARN           string        `json:"arn"`
}

type S3OwnerIdentity struct {
	PrincipalID string `json:"principalId"`
}

type S3Object struct {
	Key       string `json:"key"`
	Size      int64  `json:"size"`
	ETag      string `json:"eTag"`
	VersionID string `json:"versionId,omitempty"`
	Sequencer string `json:"sequencer"`
}

type S3GlacierEventData struct {
	RestoreEventData S3RestoreEventData `json:"restoreEventData"`
}

type S3RestoreEventData struct {
	LifecycleRestorationExpiryTime string `json:"lifecycleRestorationExpiryTime"`
	LifecycleRestoreStorageClass   string `json:"lifecycleRestoreStorageClass"`
}

type NotificationWorker struct {
	store      *db.Store
	httpClient *http.Client
	ticker     *time.Ticker
	done       chan bool
}

func NewNotificationWorker(store *db.Store) *NotificationWorker {
	return &NotificationWorker{
		store: store,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		ticker: time.NewTicker(1 * time.Second),
		done:   make(chan bool),
	}
}

func (w *NotificationWorker) Start(ctx context.Context) {
	log.Println("Notification worker started")

	for {
		select {
		case <-w.done:
			log.Println("Notification worker stopped")
			return
		case <-ctx.Done():
			log.Println("Notification worker context cancelled")
			return
		case <-w.ticker.C:
			w.processJobs(ctx)
		}
	}
}

func (w *NotificationWorker) Stop() {
	w.ticker.Stop()
	w.done <- true
}

func (w *NotificationWorker) processJobs(ctx context.Context) {
	// List all pending notification jobs
	jobs, err := w.store.Queries.ListPendingNotificationJobs(ctx)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Error listing pending notification jobs: %v", err)
		}
		return
	}

	if len(jobs) == 0 {
		return
	}

	log.Printf("Processing %d pending notification jobs", len(jobs))

	for _, job := range jobs {
		w.processJob(ctx, job)
	}
}

func (w *NotificationWorker) processJob(ctx context.Context, job db.ListPendingNotificationJobsRow) {
	// Fetch object details
	object, err := w.store.Queries.GetObjectByID(ctx, job.Event.ObjectID)
	if err != nil {
		log.Printf("Error fetching object details for job %d: %v", job.NotificationJob.ID, err)
		w.updateJobStatus(ctx, job.NotificationJob.ID, "failed", job.NotificationJob.Attempts+1, fmt.Sprintf("Failed to fetch object: %v", err))
		return
	}

	// Build the S3 event record
	eventRecord := S3EventRecord{
		EventVersion: "2.1",
		EventSource:  "aws:s3",
		AWSRegion:    "us-east-1", // Default region, could be fetched from bucket config
		EventTime:    job.Event.EventTime.Format(time.RFC3339),
		EventName:    job.Event.EventType,
		UserIdentity: S3UserIdentity{
			PrincipalID: "s3local",
		},
		RequestParameters: S3RequestParameters{
			SourceIPAddress: "127.0.0.1",
		},
		ResponseElements: S3ResponseElements{
			XAmzRequestID: fmt.Sprintf("%d", job.Event.ID),
			XAmzID2:       fmt.Sprintf("%016x", job.Event.ID),
		},
		S3: S3Entity{
			S3SchemaVersion: "1.0",
			ConfigurationID: fmt.Sprintf("%d", job.Notification.ID),
			Bucket: S3Bucket{
				Name: job.Event.BucketName,
				OwnerIdentity: S3OwnerIdentity{
					PrincipalID: "s3local",
				},
				ARN: fmt.Sprintf("arn:aws:s3:::%s", job.Event.BucketName),
			},
			Object: S3Object{
				Key:       object.Object.Key,
				Size:      object.Object.Size,
				ETag:      object.Object.ETag,
				Sequencer: fmt.Sprintf("%016x", job.Event.ID),
			},
		},
	}

	// Add versionId if present
	if object.Object.VersionID.Valid {
		eventRecord.S3.Object.VersionID = object.Object.VersionID.String
	}

	// Wrap in notification structure
	notification := S3EventNotification{
		Records: []S3EventRecord{eventRecord},
	}

	payloadBytes, err := json.Marshal(notification)
	if err != nil {
		log.Printf("Error marshalling payload for job %d: %v", job.NotificationJob.ID, err)
		w.updateJobStatus(ctx, job.NotificationJob.ID, "failed", job.NotificationJob.Attempts+1, fmt.Sprintf("Failed to marshal payload: %v", err))
		return
	}

	// Send HTTP POST request to destination ARN
	req, err := http.NewRequestWithContext(ctx, "POST", job.Notification.DestinationArn, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Error creating request for job %d: %v", job.NotificationJob.ID, err)
		w.updateJobStatus(ctx, job.NotificationJob.ID, "failed", job.NotificationJob.Attempts+1, fmt.Sprintf("Failed to create request: %v", err))
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		log.Printf("Error sending notification for job %d to %s: %v", job.NotificationJob.ID, job.Notification.DestinationArn, err)
		w.updateJobStatus(ctx, job.NotificationJob.ID, "failed", job.NotificationJob.Attempts+1, fmt.Sprintf("HTTP request failed: %v", err))
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("Successfully sent notification for job %d to %s", job.NotificationJob.ID, job.Notification.DestinationArn)
		w.updateJobStatus(ctx, job.NotificationJob.ID, "completed", job.NotificationJob.Attempts+1, "")
	} else {
		log.Printf("Failed to send notification for job %d to %s: HTTP %d", job.NotificationJob.ID, job.Notification.DestinationArn, resp.StatusCode)
		w.updateJobStatus(ctx, job.NotificationJob.ID, "failed", job.NotificationJob.Attempts+1, fmt.Sprintf("HTTP %d", resp.StatusCode))
	}
}

func (w *NotificationWorker) updateJobStatus(ctx context.Context, jobID int64, status string, attempts int64, errorMessage string) {
	var errorMsg sql.NullString
	if errorMessage != "" {
		errorMsg = sql.NullString{String: errorMessage, Valid: true}
	}

	err := w.store.Queries.UpdateNotificationJobStatus(ctx, db.UpdateNotificationJobStatusParams{
		Status:       status,
		Attempts:     attempts,
		ErrorMessage: errorMsg,
		ID:           jobID,
	})
	if err != nil {
		log.Printf("Error updating job status for job %d: %v", jobID, err)
	}
}
