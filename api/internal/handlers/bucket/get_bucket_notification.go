package bucket

import (
	"encoding/xml"
	"net/http"

	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/handlers/s3error"
)

// GetBucketNotificationConfiguration handles GET /{bucket}?notification
func GetBucketNotificationConfiguration(w http.ResponseWriter, r *http.Request) {
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

	// Get all notifications for this bucket
	notifications, err := store.Queries.ListNotificationsByBucket(r.Context(), bucketName)
	if err != nil {
		s3error.NewInternalError(err).WriteError(w)
		return
	}

	// Build notification configuration response
	config := NotificationConfiguration{}

	// Group notifications by destination type and ARN
	queueConfigs := make(map[string]*QueueConfiguration)
	topicConfigs := make(map[string]*TopicConfiguration)
	lambdaConfigs := make(map[string]*LambdaFunctionConfiguration)

	for _, notification := range notifications {
		if !notification.Enabled {
			continue
		}

		// Build filter rules
		var filterRules []FilterRule
		if notification.FilterPrefix.Valid {
			filterRules = append(filterRules, FilterRule{
				Name:  "prefix",
				Value: notification.FilterPrefix.String,
			})
		}
		if notification.FilterSuffix.Valid {
			filterRules = append(filterRules, FilterRule{
				Name:  "suffix",
				Value: notification.FilterSuffix.String,
			})
		}

		filter := NotificationFilter{
			S3Key: S3KeyFilter{
				FilterRules: filterRules,
			},
		}

		switch notification.DestinationType {
		case "sqs":
			if queueConfig, exists := queueConfigs[notification.DestinationArn]; exists {
				queueConfig.Events = append(queueConfig.Events, notification.EventType)
			} else {
				queueConfigs[notification.DestinationArn] = &QueueConfiguration{
					Queue:  notification.DestinationArn,
					Events: []string{notification.EventType},
					Filter: filter,
				}
			}
		case "sns":
			if topicConfig, exists := topicConfigs[notification.DestinationArn]; exists {
				topicConfig.Events = append(topicConfig.Events, notification.EventType)
			} else {
				topicConfigs[notification.DestinationArn] = &TopicConfiguration{
					Topic:  notification.DestinationArn,
					Events: []string{notification.EventType},
					Filter: filter,
				}
			}
		case "lambda":
			if lambdaConfig, exists := lambdaConfigs[notification.DestinationArn]; exists {
				lambdaConfig.Events = append(lambdaConfig.Events, notification.EventType)
			} else {
				lambdaConfigs[notification.DestinationArn] = &LambdaFunctionConfiguration{
					LambdaFunctionArn: notification.DestinationArn,
					Events:            []string{notification.EventType},
					Filter:            filter,
				}
			}
		}
	}

	// Convert maps to slices
	for _, queueConfig := range queueConfigs {
		config.QueueConfigurations = append(config.QueueConfigurations, *queueConfig)
	}
	for _, topicConfig := range topicConfigs {
		config.TopicConfigurations = append(config.TopicConfigurations, *topicConfig)
	}
	for _, lambdaConfig := range lambdaConfigs {
		config.LambdaFunctionConfigurations = append(config.LambdaFunctionConfigurations, *lambdaConfig)
	}

	// Marshal and write response
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)

	xmlData, err := xml.MarshalIndent(config, "", "  ")
	if err != nil {
		s3error.NewInternalError(err).WriteError(w)
		return
	}

	w.Write([]byte(xml.Header))
	w.Write(xmlData)
}
