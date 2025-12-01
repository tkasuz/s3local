package bucket

import (
	"database/sql"
	"encoding/xml"
	"io"
	"net/http"

	"github.com/tkasuz/s3local/internal/db"
	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/handlers/s3error"
)

// PutBucketNotificationConfiguration handles PUT /{bucket}?notification
func PutBucketNotificationConfiguration(w http.ResponseWriter, r *http.Request) {
	store := ctx.GetStore(r.Context())
	bucketName := ctx.GetBucketName(r.Context())

	// Check if bucket exists
	exists, err := store.Queries.BucketExists(r.Context(), bucketName)
	if err != nil {
		s3error.NewInternalError(err).WriteError(w)
		return
	}
	if !exists {
		s3error.NewNoSuchBucketError(bucketName).WriteError(w)
		return
	}

	// Parse XML body
	if r.Body == nil || r.ContentLength == 0 {
		s3error.NewMalformedXMLError().WriteError(w)
		return
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s3error.NewInternalError(err).WriteError(w)
		return
	}

	var notificationConfig NotificationConfiguration
	if err := xml.Unmarshal(body, &notificationConfig); err != nil {
		s3error.NewMalformedXMLError().WriteError(w)
		return
	}

	// Delete existing notifications for this bucket
	// Get existing notifications first
	existingNotifications, err := store.Queries.ListNotificationsByBucket(r.Context(), bucketName)
	if err != nil {
		s3error.NewInternalError(err).WriteError(w)
		return
	}

	// Delete all existing notifications
	for _, notification := range existingNotifications {
		if err := store.Queries.DeleteNotification(r.Context(), notification.ID); err != nil {
			s3error.NewInternalError(err).WriteError(w)
			return
		}
	}

	// Insert new QueueConfigurations
	for _, queueConfig := range notificationConfig.QueueConfigurations {
		for _, event := range queueConfig.Events {
			var filterPrefix sql.NullString
			var filterSuffix sql.NullString

			if queueConfig.Filter.S3Key.FilterRules != nil {
				for _, rule := range queueConfig.Filter.S3Key.FilterRules {
					if rule.Name == "prefix" {
						filterPrefix = sql.NullString{String: rule.Value, Valid: true}
					}
					if rule.Name == "suffix" {
						filterSuffix = sql.NullString{String: rule.Value, Valid: true}
					}
				}
			}

			_, err := store.Queries.CreateNotification(r.Context(), db.CreateNotificationParams{
				BucketName:      bucketName,
				EventType:       event,
				DestinationType: "sqs",
				DestinationArn:  queueConfig.Queue,
				FilterPrefix:    filterPrefix,
				FilterSuffix:    filterSuffix,
				Enabled:         true,
			})
			if err != nil {
				s3error.NewInternalError(err).WriteError(w)
				return
			}
		}
	}

	// Insert new TopicConfigurations
	for _, topicConfig := range notificationConfig.TopicConfigurations {
		for _, event := range topicConfig.Events {
			var filterPrefix sql.NullString
			var filterSuffix sql.NullString

			if topicConfig.Filter.S3Key.FilterRules != nil {
				for _, rule := range topicConfig.Filter.S3Key.FilterRules {
					if rule.Name == "prefix" {
						filterPrefix = sql.NullString{String: rule.Value, Valid: true}
					}
					if rule.Name == "suffix" {
						filterSuffix = sql.NullString{String: rule.Value, Valid: true}
					}
				}
			}

			_, err := store.Queries.CreateNotification(r.Context(), db.CreateNotificationParams{
				BucketName:      bucketName,
				EventType:       event,
				DestinationType: "sns",
				DestinationArn:  topicConfig.Topic,
				FilterPrefix:    filterPrefix,
				FilterSuffix:    filterSuffix,
				Enabled:         true,
			})
			if err != nil {
				s3error.NewInternalError(err).WriteError(w)
				return
			}
		}
	}

	// Insert new LambdaFunctionConfigurations
	for _, lambdaConfig := range notificationConfig.LambdaFunctionConfigurations {
		for _, event := range lambdaConfig.Events {
			var filterPrefix sql.NullString
			var filterSuffix sql.NullString

			if lambdaConfig.Filter.S3Key.FilterRules != nil {
				for _, rule := range lambdaConfig.Filter.S3Key.FilterRules {
					if rule.Name == "prefix" {
						filterPrefix = sql.NullString{String: rule.Value, Valid: true}
					}
					if rule.Name == "suffix" {
						filterSuffix = sql.NullString{String: rule.Value, Valid: true}
					}
				}
			}

			_, err := store.Queries.CreateNotification(r.Context(), db.CreateNotificationParams{
				BucketName:      bucketName,
				EventType:       event,
				DestinationType: "lambda",
				DestinationArn:  lambdaConfig.LambdaFunctionArn,
				FilterPrefix:    filterPrefix,
				FilterSuffix:    filterSuffix,
				Enabled:         true,
			})
			if err != nil {
				s3error.NewInternalError(err).WriteError(w)
				return
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}

// NotificationConfiguration represents the S3 notification configuration XML structure
type NotificationConfiguration struct {
	XMLName                      struct{}                      `xml:"NotificationConfiguration"`
	QueueConfigurations          []QueueConfiguration          `xml:"QueueConfiguration"`
	TopicConfigurations          []TopicConfiguration          `xml:"TopicConfiguration"`
	LambdaFunctionConfigurations []LambdaFunctionConfiguration `xml:"CloudFunctionConfiguration"`
}

type QueueConfiguration struct {
	Id     string             `xml:"Id,omitempty"`
	Queue  string             `xml:"Queue"`
	Events []string           `xml:"Event"`
	Filter NotificationFilter `xml:"Filter,omitempty"`
}

type TopicConfiguration struct {
	Id     string             `xml:"Id,omitempty"`
	Topic  string             `xml:"Topic"`
	Events []string           `xml:"Event"`
	Filter NotificationFilter `xml:"Filter,omitempty"`
}

type LambdaFunctionConfiguration struct {
	Id                string             `xml:"Id,omitempty"`
	LambdaFunctionArn string             `xml:"CloudFunction"`
	Events            []string           `xml:"Event"`
	Filter            NotificationFilter `xml:"Filter,omitempty"`
}

type NotificationFilter struct {
	S3Key S3KeyFilter `xml:"S3Key"`
}

type S3KeyFilter struct {
	FilterRules []FilterRule `xml:"FilterRule"`
}

type FilterRule struct {
	Name  string `xml:"Name"`
	Value string `xml:"Value"`
}
