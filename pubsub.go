package pubsub

import (
	"encoding/json"
	"fmt"

	"github.com/labstack/gommon/log"
	"golang.org/x/net/websocket"
)

const (
	PUBLISH     = "publish"
	SUBSCRIBE   = "subscribe"
	UNSUBSCRIBE = "unsubscribe"
)

type Broker struct {
	Clients       []Client
	Subscriptions []Subscription
}

type Client struct {
	Name       string
	ID         string
	Connection *websocket.Conn
}

type Message struct {
	Action  string          `json:"action"`
	Topic   string          `json:"topic"`
	Message json.RawMessage `json:"message"`
}

type Subscription struct {
	Topic  string
	Client *Client
}

func (ps *Broker) AddClient(client Client) {

	ps.Clients = append(ps.Clients, client)
}

func (ps *Broker) RemoveClient(client Client) *Broker {

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

func (ps *Broker) GetSubscriptions(topic string, client *Client) []Subscription {

	var subscriptionList []Subscription

	for _, subscription := range ps.Subscriptions {

		if client != nil {

			if subscription.Client.ID == client.ID && subscription.Topic == topic {
				subscriptionList = append(subscriptionList, subscription)

			}
		} else {

			if subscription.Topic == topic {
				subscriptionList = append(subscriptionList, subscription)
			}
		}
	}

	return subscriptionList
}

func (ps *Broker) Subscribe(client *Client, topic string) *Broker {

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

func (ps *Broker) Publish(topic string, message []byte, excludeClient *Client) {

	subscriptions := ps.GetSubscriptions(topic, nil)

	for _, sub := range subscriptions {

		fmt.Printf("Sending to client id %s message is %s \n", sub.Client.ID, message)
		err := sub.Client.Send(message)
		if err != nil {
			log.Errorf("error sending message to client: %w", err)
		}
	}

}
func (client *Client) Send(message []byte) error {
	return websocket.Message.Send(client.Connection, message)
}

func (ps *Broker) Unsubscribe(client *Client, topic string) *Broker {

	//clientSubscriptions := ps.GetSubscriptions(topic, client)
	for index, sub := range ps.Subscriptions {

		if sub.Client.ID == client.ID && sub.Topic == topic {
			// found this subscription from client and we do need remove it
			ps.Subscriptions = append(ps.Subscriptions[:index], ps.Subscriptions[index+1:]...)
		}
	}

	return ps
}

func (ps *Broker) HandleReceiveMessage(client Client, payload []byte) *Broker {
	m := Message{}

	err := json.Unmarshal(payload, &m)
	if err != nil {
		fmt.Println("This is not correct message payload")
		return ps
	}

	fmt.Printf("handle message %s from %s", m.Action, client.Name)
	switch m.Action {
	case PUBLISH:
		fmt.Println("publish message")
		ps.Publish(m.Topic, m.Message, nil)
		break
	case SUBSCRIBE:
		ps.Subscribe(&client, m.Topic)
		fmt.Println("new subscriber to topic", m.Topic, len(ps.Subscriptions), client.ID)
		break
	case UNSUBSCRIBE:
		fmt.Println("Client want to unsubscribe the topic", m.Topic, client.ID)
		ps.Unsubscribe(&client, m.Topic)
		break
	default:
		break
	}

	return ps
}
