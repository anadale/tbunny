package rmq

import (
	"encoding/json"
	"fmt"
	"net/url"

	rabbithole "github.com/michaelklishin/rabbit-hole/v3"
)

// Mainly, we need this replacement (duplication) because RabbitMQ **sometimes** returns
// an empty array instead of null for the `owner_pid_details` field. Since we can't
// fix RabbitMQ, we need to handle this case here.
// https://github.com/rabbitmq/rabbitmq-server/discussions/15734

type DetailedQueueInfo struct {
	// Queue name
	Name string `json:"name"`
	// Queue type
	Type string `json:"type,omitempty"`
	// Virtual host this queue belongs to
	Vhost string `json:"vhost,omitempty"`
	// Is this queue durable?
	Durable bool `json:"durable"`
	// Is this queue auto-deleted?
	AutoDelete rabbithole.AutoDelete `json:"auto_delete"`
	// Is this queue exclusive?
	Exclusive bool `json:"exclusive,omitempty"`
	// Extra queue arguments
	Arguments map[string]interface{} `json:"arguments"`

	// RabbitMQ node that hosts the leader replica for this queue
	Node string `json:"node,omitempty"`
	// Queue status
	Status string `json:"state,omitempty"`
	// Queue leader when it is a quorum queue
	Leader string `json:"leader,omitempty"`
	// Queue members when it is a quorum queue
	Members []string `json:"members,omitempty"`
	// Queue online members when it is a quorum queue
	Online []string `json:"online,omitempty"`

	// Total amount of RAM used by this queue
	Memory int64 `json:"memory,omitempty"`
	// How many consumers this queue has
	Consumers int `json:"consumers,omitempty"`
	// Detail information of consumers
	ConsumerDetails *[]rabbithole.ConsumerDetail `json:"consumer_details,omitempty"`
	// Utilisation of all the consumers
	ConsumerUtilisation float64 `json:"consumer_utilisation,omitempty"`
	// If there is an exclusive consumer, its consumer tag
	ExclusiveConsumerTag string `json:"exclusive_consumer_tag,omitempty"`

	// GarbageCollection metrics
	GarbageCollection *rabbithole.GarbageCollectionDetails `json:"garbage_collection,omitempty"`

	// Policy applied to this queue, if any
	Policy string `json:"policy,omitempty"`

	// Total bytes of messages in this queue
	MessagesBytes               int64 `json:"message_bytes,omitempty"`
	MessagesBytesPersistent     int64 `json:"message_bytes_persistent,omitempty"`
	MessagesBytesRAM            int64 `json:"message_bytes_ram,omitempty"`
	MessagesBytesReady          int64 `json:"message_bytes_ready,omitempty"`
	MessagesBytesUnacknowledged int64 `json:"message_bytes_unacknowledged,omitempty"`

	// Total number of messages in this queue
	Messages           int                     `json:"messages,omitempty"`
	MessagesDetails    *rabbithole.RateDetails `json:"messages_details,omitempty"`
	MessagesPersistent int                     `json:"messages_persistent,omitempty"`
	MessagesRAM        int                     `json:"messages_ram,omitempty"`

	// Number of messages ready to be delivered
	MessagesReady        int                     `json:"messages_ready,omitempty"`
	MessagesReadyDetails *rabbithole.RateDetails `json:"messages_ready_details,omitempty"`

	// Number of messages delivered and pending acknowledgements from consumers
	MessagesUnacknowledged        int                     `json:"messages_unacknowledged,omitempty"`
	MessagesUnacknowledgedDetails *rabbithole.RateDetails `json:"messages_unacknowledged_details,omitempty"`

	MessageStats *rabbithole.MessageStats `json:"message_stats,omitempty"`

	OwnerPidDetails *OwnerPidDetailsWrapper `json:"owner_pid_details,omitempty"`

	BackingQueueStatus *rabbithole.BackingQueueStatus `json:"backing_queue_status,omitempty"`

	ActiveConsumers int64 `json:"active_consumers,omitempty"`
}

type OwnerPidDetailsWrapper struct {
	*rabbithole.OwnerPidDetails
}

// GetQueue returns information about a queue.
func (c *Client) GetQueue(vhost, queue string) (rec *DetailedQueueInfo, err error) {
	req, err := newGETRequest(c, "queues/"+url.PathEscape(vhost)+"/"+url.PathEscape(queue))
	if err != nil {
		return nil, err
	}

	if err = executeAndParseRequest(c, req, &rec); err != nil {
		return nil, err
	}

	return rec, nil
}

func (w *OwnerPidDetailsWrapper) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		w.OwnerPidDetails = nil
		return nil
	}

	switch data[0] {
	case '[': // It can only be an empty array, so we don't care about the rest
		w.OwnerPidDetails = nil
		return nil
	case '{':
		var v rabbithole.OwnerPidDetails
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}

		w.OwnerPidDetails = &v
		return nil
	}

	return fmt.Errorf("unexpected JSON of OwnerPidDetails: %s", data)
}
