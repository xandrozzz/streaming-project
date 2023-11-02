package streaming

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nats-io/nats.go"
	"log"
	"os"
	"streaming-project/database"
	"streaming-project/https"
	"streaming-project/models"
	"time"
)

func connectToNATS() (*nats.Conn, error) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}

	return nc, nil
}

func natsJetStream(nc *nats.Conn) (nats.JetStreamContext, error) {
	jsCtx, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("jetstream: %w", err)
	}

	return jsCtx, nil
}

func createStream(ctx context.Context, jsCtx nats.JetStreamContext) (*nats.StreamInfo, error) {

	streamName, exists := os.LookupEnv("STREAM_NAME")
	if !exists {
		log.Fatalln("NATS stream name not found in .env file!")
	}

	stream, _ := jsCtx.StreamInfo(streamName)

	var err error
	if stream == nil {

		stream, err = jsCtx.AddStream(&nats.StreamConfig{
			Name:              streamName,
			Subjects:          []string{streamName + ".*"},
			Retention:         nats.WorkQueuePolicy,
			Discard:           nats.DiscardOld,
			MaxAge:            7 * 24 * time.Hour,
			Storage:           nats.FileStorage,
			MaxMsgsPerSubject: 100_000_000,
			MaxMsgSize:        4 << 20,
			NoAck:             false,
		}, nats.Context(ctx))
		if err != nil {
			return nil, fmt.Errorf("add stream: %w", err)
		}
	}

	return stream, nil
}

func createConsumer(ctx context.Context, jsCtx nats.JetStreamContext, consumerGroupName, streamName string) (*nats.ConsumerInfo, error) {
	consumer, err := jsCtx.AddConsumer(streamName, &nats.ConsumerConfig{
		Durable:       consumerGroupName,
		DeliverPolicy: nats.DeliverAllPolicy,
		AckPolicy:     nats.AckExplicitPolicy,
		AckWait:       10 * time.Second,
		MaxAckPending: -1,
	}, nats.Context(ctx))
	if err != nil {
		return nil, fmt.Errorf("add consumer: %w", err)
	}

	return consumer, nil
}

func subscribe(ctx context.Context, js nats.JetStreamContext, subject, consumerGroupName, streamName string, db *database.Database) error {
	pullSub, err := js.PullSubscribe(
		subject,
		consumerGroupName,
		nats.ManualAck(),
		nats.Bind(streamName, consumerGroupName),
		nats.Context(ctx),
	)
	if err != nil {
		return fmt.Errorf("pull subscribe: %w", err)
	}

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for range ticker.C {
			con, err := pullSub.ConsumerInfo()
			if err != nil {
				log.Fatalln(err)
			}
			log.Printf("Consumer data: NumPending=%v NumWaiting=%v\n", con.NumPending, con.NumWaiting)
		}
	}()

	go func() {
		for {
			msgs, err := pullSub.Fetch(1, nats.MaxWait(30*time.Second))
			select {
			case <-ctx.Done():
				fmt.Println("Context is finished, exiting")
				return
			default:
			}
			if err != nil {
				if errors.Is(err, nats.ErrTimeout) {
					continue
				}
				log.Fatalln("fetch failed", err)
			}
			for _, msg := range msgs {
				var order models.Order

				if !json.Valid(msg.Data) {
					log.Println("Got invalid data, skipping")
					continue
				}

				err := json.Unmarshal(msg.Data, &order)
				if err != nil {
					log.Fatalln(err)
				}
				log.Println("Adding message to the database")
				err = db.AddOrder(&order)
				if err != nil {
					log.Fatalln(err)
				}

				err = msg.Ack()
				if err != nil {
					log.Fatalln("Cannot ack message")
				}
			}
		}
	}()

	return nil
}

func StartConsumer(db *database.Database) error {
	conn, err := connectToNATS()
	if err != nil {
		return err
	}
	defer func(conn *nats.Conn) {
		err := conn.Drain()
		if err != nil {
			log.Fatalln(err)
		}
	}(conn)

	jsctx, err := natsJetStream(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := createStream(ctx, jsctx)
	if err != nil {
		return err
	}

	consumerName, exists := os.LookupEnv("DURABLE_CONSUMER_NAME")
	if !exists {
		log.Fatalln("Durable consumer name not found in .env file!")
	}

	var consumer *nats.ConsumerInfo
	consumer, err = createConsumer(ctx, jsctx, consumerName, stream.Config.Name)
	if err != nil {
		return err
	}

	defer func(jsctx nats.JetStreamContext, stream, consumer string, cancel func()) {
		cancel()
		err := jsctx.DeleteConsumer(stream, consumer)
		if err != nil {
			log.Fatalln(err)
		}
	}(jsctx, stream.Config.Name, consumer.Name, cancel)

	consumeSubject, exists := os.LookupEnv("CONSUME_SUBJECT")
	if !exists {
		log.Fatalln("Subject for consumption not found in .env file!")
	}

	err = subscribe(ctx, jsctx, consumeSubject, consumer.Name, stream.Config.Name, db)
	if err != nil {
		return err
	}

	routerPort, exists := os.LookupEnv("ROUTER_PORT")
	if !exists {
		log.Fatalln("Port for http server not found in .env file!")
	}
	router := https.NewRouter(routerPort, db)
	err = router.Start()
	if err != nil {
		return err
	}

	return nil
}
