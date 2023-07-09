package pubsub

import (
	"encoding/json"

	"github.com/jtbonhomme/pubsub/client"
	"github.com/rs/zerolog"
)

const (
	PUBLISH     = "publish"     // Publish a message.
	SUBSCRIBE   = "subscribe"   // Subscribe to a topic.
	UNSUBSCRIBE = "unsubscribe" // Unsubscribe from a topic.
)

// Broker is a pubsub service message broker.
type Broker struct {
	log           *zerolog.Logger
	Clients       []*client.Client // List of active clients of the broker.
	Subscriptions []Subscription   // List of topic subscriptions.
}

// Message is describe a communication message through a topic in the pubsub service.
type Message struct {
	Action  string          `json:"action"`  // Action can be `PUBLISH`, `SUBSCRIBE` or `UNSUBSCRIBE`.
	Topic   string          `json:"topic"`   // Topic is a pubsub service topic name.
	Payload json.RawMessage `json:"message"` // Payload is sent in the topic in case action is `PUBLISH`.
}

// Subscription represents a pubsub service client subscribing to a topic.
type Subscription struct {
	Topic  string         // Topic is the name of the topic the client subscribes to.
	Client *client.Client // Client is a pubsub service client.
}

// New instantiates a new pubsub service broker.
func New(logger *zerolog.Logger) *Broker {
	return &Broker{
		log: logger,
	}
}

// AddClient registers a client to the pubsub service.
func (ps *Broker) AddClient(client *client.Client) {
	ps.Clients = append(ps.Clients, client)
}

// RmoveClient unregisters a client from the pubsub service.
func (ps *Broker) RemoveClient(client *client.Client) *Broker {
	// first remove all subscriptions by this client
	for index, sub := range ps.Subscriptions {
		if client.ID == sub.Client.ID {
			ps.Subscriptions = append(ps.Subscriptions[:index], ps.Subscriptions[index+1:]...)
		}
	}

	// remove client from the list
	for index, c := range ps.Clients {
		if c.ID == client.ID {
			ps.Clients = append(ps.Clients[:index], ps.Clients[index+1:]...)
		}
	}

	return ps
}

// GetSubscriptions lists active subscriptions for a topic or a client.
func (ps *Broker) GetSubscriptions(topic string, client *client.Client) []Subscription {
	var subscriptionList []Subscription

	for _, subscription := range ps.Subscriptions {
		if client != nil {
			if subscription.Client.ID == client.ID && subscription.Topic == topic {
				subscriptionList = append(subscriptionList, subscription)

			}
		} else if subscription.Topic == topic {
			subscriptionList = append(subscriptionList, subscription)
		}
	}

	return subscriptionList
}

// Subscribe adds a subscription from a client to a topic.
func (ps *Broker) Subscribe(client *client.Client, topic string) *Broker {
	clientSubs := ps.GetSubscriptions(topic, client)

	if len(clientSubs) > 0 {
		// client is subscribed this topic before
		return ps
	}

	newSubscription := Subscription{
		Topic:  topic,
		Client: client,
	}

	ps.Subscriptions = append(ps.Subscriptions, newSubscription)

	return ps
}

// Publish sends a message in a topic.
// Clients can be excluded from receiving the message.
func (ps *Broker) Publish(topic string, message []byte, excludeClient *client.Client) {
	subscriptions := ps.GetSubscriptions(topic, nil)

	for _, sub := range subscriptions {

		ps.log.Debug().Msgf("Sending to client id %s message is %s \n", sub.Client.ID, message)
		err := sub.Client.Send(message)
		if err != nil {
			ps.log.Error().Msgf("error sending message to client: %w", err)
		}
	}
}

// Unsubscribe removes a client's subscription from a topic.
func (ps *Broker) Unsubscribe(client *client.Client, topic string) *Broker {
	//clientSubscriptions := ps.GetSubscriptions(topic, client)
	for index, sub := range ps.Subscriptions {

		if sub.Client.ID == client.ID && sub.Topic == topic {
			// found this subscription from client and we do need remove it
			ps.Subscriptions = append(ps.Subscriptions[:index], ps.Subscriptions[index+1:]...)
		}
	}

	return ps
}

// HandleReceiveMessage executes action of a receive Message.
func (ps *Broker) HandleReceiveMessage(client *client.Client, payload []byte) *Broker {
	m := Message{}

	err := json.Unmarshal(payload, &m)
	if err != nil {
		ps.log.Debug().Msg("This is not correct message payload")
		return ps
	}

	ps.log.Debug().Msgf("handle message %s from %s", m.Action, client.Name)
	switch m.Action {
	case PUBLISH:
		ps.log.Debug().Msg("publish message")
		ps.Publish(m.Topic, m.Payload, nil)
		break
	case SUBSCRIBE:
		ps.Subscribe(client, m.Topic)
		ps.log.Debug().Msgf("client %s (%s) subscribes to topic %s", client.Name, client.ID, m.Topic)
		break
	case UNSUBSCRIBE:
		ps.log.Debug().Msgf("client %s (%s) unsubscribes from topic", client.Name, client.ID, m.Topic)
		ps.Unsubscribe(client, m.Topic)
		break
	default:
		break
	}

	return ps
}
