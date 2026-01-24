// adapter/event/pubsub/topic_resolver.go
package pubsub

import (
	domainevent "github.com/histopathai/main-service/internal/domain/event"
)

type TopicResolver struct {
	topicMapping map[domainevent.EventType]string
}

func NewTopicResolver(topicMapping map[domainevent.EventType]string) *TopicResolver {
	return &TopicResolver{
		topicMapping: topicMapping,
	}
}

func (r *TopicResolver) ResolveTopic(eventType domainevent.EventType) string {
	if topic, ok := r.topicMapping[eventType]; ok {
		return topic
	}
	return "default-topic"
}
