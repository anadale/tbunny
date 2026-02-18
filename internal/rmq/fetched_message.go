package rmq

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
)

type FetchedMessage struct {
	PayloadBytes    int                      `json:"payload_bytes,omitempty"`
	Redelivered     bool                     `json:"redelivered,omitempty"`
	Exchange        string                   `json:"exchange,omitempty"`
	RoutingKey      string                   `json:"routing_key,omitempty"`
	MessageCount    int                      `json:"message_count,omitempty"`
	Properties      FetchedMessageProperties `json:"properties,omitempty"`
	Payload         string                   `json:"payload"`
	PayloadEncoding PayloadEncoding          `json:"payload_encoding"`
}

type FetchedMessageProperties struct {
	AppId           string              `json:"app_id,omitempty"`
	ContentEncoding string              `json:"content_encoding,omitempty"`
	ContentType     string              `json:"content_type,omitempty"`
	CorrelationId   string              `json:"correlation_id,omitempty"`
	DeliveryMode    MessageDeliveryMode `json:"delivery_mode,omitempty"`
	Expiration      string              `json:"expiration,omitempty"`
	Headers         map[string]any      `json:"headers,omitempty"`
	MessageId       string              `json:"message_id,omitempty"`
	Priority        int                 `json:"priority,omitempty"`
	ReplyTo         string              `json:"reply_to,omitempty"`
	Timestamp       int64               `json:"timestamp,omitempty"`
	Type            string              `json:"type,omitempty"`
	UserId          string              `json:"user_id,omitempty"`
}

type AckMode string
type MessageDeliveryMode int
type PayloadEncoding string
type RequestedMessageEncoding string

const (
	AckModeAckRequeueTrue            AckMode                  = "ack_requeue_true"
	AckModeAckRequeueFalse           AckMode                  = "ack_requeue_false"
	AckModeRejectRequeueTrue         AckMode                  = "reject_requeue_true"
	AckModeRejectRequeueFalse        AckMode                  = "reject_requeue_false"
	PayloadEncodingBase64            PayloadEncoding          = "base64"
	PayloadEncodingString            PayloadEncoding          = "string"
	RequestedMessageEncodingBase64   RequestedMessageEncoding = "base64"
	RequestedMessageEncodingAuto     RequestedMessageEncoding = "auto"
	MessageDeliveryModeUnknown       MessageDeliveryMode      = 0
	MessageDeliveryModeNonPersistent MessageDeliveryMode      = 1
	MessageDeliveryModePersistent    MessageDeliveryMode      = 2
)

func (d MessageDeliveryMode) String() string {
	switch d {
	case MessageDeliveryModeNonPersistent:
		return "NonPersistent"
	case MessageDeliveryModePersistent:
		return "Persistent"
	default:
		return "Unknown"
	}
}

func ParseDeliveryMode(s string) (MessageDeliveryMode, error) {
	switch s {
	case "NonPersistent":
		return MessageDeliveryModeNonPersistent, nil
	case "Persistent":
		return MessageDeliveryModePersistent, nil
	default:
		return MessageDeliveryModeUnknown, fmt.Errorf("unknown delivery mode: %s", s)
	}
}

type fetchMessagesRequest struct {
	Count    int                      `json:"count"`
	AckMode  AckMode                  `json:"ackmode"`
	Encoding RequestedMessageEncoding `json:"encoding"`
}

func (c *Client) GetQueueMessages(vhost, queue string, ackMode AckMode, encoding RequestedMessageEncoding, count int) (messages []*FetchedMessage, err error) {
	if ackMode != AckModeAckRequeueTrue && ackMode != AckModeAckRequeueFalse &&
		ackMode != AckModeRejectRequeueTrue && ackMode != AckModeRejectRequeueFalse {
		return nil, errors.New("invalid ack mode")
	}
	if encoding != RequestedMessageEncodingBase64 && encoding != RequestedMessageEncodingAuto {
		return nil, errors.New("invalid message encoding")
	}
	if count <= 0 {
		return nil, errors.New("count must be positive")
	}

	fetchRequest := fetchMessagesRequest{
		Count:    count,
		AckMode:  ackMode,
		Encoding: encoding,
	}

	body, err := json.Marshal(fetchRequest)
	if err != nil {
		return nil, err
	}

	req, err := newRequestWithBody(c, "POST", "queues/"+url.PathEscape(vhost)+"/"+url.PathEscape(queue)+"/get", body)
	if err != nil {
		return nil, err
	}

	if err = executeAndParseRequest(c, req, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}
