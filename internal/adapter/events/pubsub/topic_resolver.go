// adapter/event/pubsub/topic_resolver.go
package pubsub

import (
	domainevent "github.com/histopathai/main-service/internal/domain/event"
)

type TopicResolver struct {
	topicMapping map[domainevent.EventType]string
}

func NewTopicResolver(projectID string) *TopicResolver {
	return &TopicResolver{
		topicMapping: map[domainevent.EventType]string{
			domainevent.UploadedEventType:                 "image-uploaded",
			domainevent.DeleteEventType:                   "image-deleted",
			domainevent.ImageProcessingRequestedEventType: "image-processing-requested",
			domainevent.ImageProcessingCompletedEventType: "image-processing-completed",
		},
	}
}

func (r *TopicResolver) ResolveTopic(eventType domainevent.EventType) string {
	if topic, ok := r.topicMapping[eventType]; ok {
		return topic
	}
	return "default-topic"
}
