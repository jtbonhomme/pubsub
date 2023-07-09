package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/goombaio/namegenerator"
	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/net/websocket"

	"github.com/rs/zerolog"

	"github.com/jtbonhomme/pubsub"
	"github.com/jtbonhomme/pubsub/client"
)

const skipFrameCount = 3

var ps *pubsub.Broker

var nameGenerator namegenerator.Generator

func hello(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()

		client := client.New(
			nameGenerator.Generate(),
			ws,
		)

		// add this client into the list
		ps.AddClient(client)
		log.Printf("New Client %s - %s is connected, total clients %d and subscriptions %d", client.Name, client.ID, len(ps.Clients), len(ps.Subscriptions))

		for {
			var err error
			// Read
			var msg = make([]byte, 512)
			err = websocket.Message.Receive(ws, &msg)
			if err != nil {
				ps.RemoveClient(client)
				c.Logger().Error(err)
				log.Printf("removed client %s total clients %d and subscriptions %d", client.Name, len(ps.Clients), len(ps.Subscriptions))
				return
			}

			ps.HandleReceiveMessage(client, msg)
		}
	}).ServeHTTP(c.Response(), c.Request())

	return nil
}

func main() {
	seed := time.Now().UTC().UnixNano()
	nameGenerator = namegenerator.NewNameGenerator(seed)
	var l zerolog.Level

	switch strings.ToLower("debug") {
	case "error":
		l = zerolog.ErrorLevel
	case "warn":
		l = zerolog.WarnLevel
	case "info":
		l = zerolog.InfoLevel
	case "debug":
		l = zerolog.DebugLevel
	default:
		l = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(l)

	output := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}
	logger := zerolog.New(output).With().Timestamp().CallerWithSkipFrameCount(zerolog.CallerSkipFrameCount + skipFrameCount).Logger()

	ps = pubsub.New(&logger)

	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))
	e.Use(middleware.Recover())
	e.Static("/", "panel")
	e.GET("/ws", hello)

	e.Logger.Fatal(e.Start(":12345"))
}
