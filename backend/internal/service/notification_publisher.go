package service

type multiNotificationPublisher struct {
	publishers []NotificationPublisher
}

func NewMultiNotificationPublisher(publishers ...NotificationPublisher) NotificationPublisher {
	return &multiNotificationPublisher{publishers: publishers}
}

func (p *multiNotificationPublisher) PublishToUser(userID int64, topic string, payload interface{}) bool {
	delivered := false
	for _, publisher := range p.publishers {
		if publisher != nil && publisher.PublishToUser(userID, topic, payload) {
			delivered = true
		}
	}
	return delivered
}
