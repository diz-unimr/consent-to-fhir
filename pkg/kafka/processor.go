package kafka

import (
	"consent-to-fhir/pkg/config"
	"consent-to-fhir/pkg/mapper"
	cKafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
)

type Processor struct {
	config config.AppConfig
	mapper *mapper.GicsMapper
}

func NewProcessor(config config.AppConfig) *Processor {
	return &Processor{
		config: config,
		mapper: mapper.NewGicsMapper(config),
	}
}

func (p *Processor) Run() {
	// signal handler to break the loop
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// create producer
	producer := NewProducer(p.config.Kafka)
	var wg sync.WaitGroup

	for i := 1; i <= p.config.Kafka.NumConsumers; i++ {

		wg.Add(1)

		go func(clientId string) {

			// create consumer with subscription to input topic
			c := NewConsumer(p.config, clientId)
			log.WithFields(log.Fields{
				"group-id":  p.config.App.Name,
				"client-id": clientId,
			}).
				Info("Consumer created")

			for {
				select {
				case <-sigchan:
					log.WithField("client-id", c.ClientId).Info("Consumer shutting down gracefully")
					syncConsumerCommits(c)
					wg.Done()
					return

				default:
					msg, err := c.Consumer.ReadMessage(1000)
					if err == nil {
						log.WithFields(log.Fields{
							"client-id": clientId,
							"key":       string(msg.Key),
							"topic":     *msg.TopicPartition.Topic,
							"offset":    msg.TopicPartition.Offset.String()}).
							Debug("Message received")

						deliveryChan := createListener(sigchan, c, msg)
						p.processMessages(producer, msg, deliveryChan, sigchan)

					} else {
						if err.(cKafka.Error).Code() != cKafka.ErrTimedOut {
							// The producer will automatically try to recover from all errors.
							log.WithError(err).Error("Consumer error")
						}
					}
				}
			}
		}(strconv.Itoa(i))
	}
	<-sigchan
	close(sigchan)
	wg.Wait()
	log.Info("All consumers stopped. Flushing outstanding producer messages...")

	for producer.Producer.Flush(10000) > 0 {
		log.Debug("Still waiting to flush outstanding messages")
	}
	log.Info("Done")
	producer.Producer.Close()
}

func createListener(sigchan chan os.Signal, c *ConsentConsumer, msg *cKafka.Message) chan cKafka.Event {
	listener := make(chan cKafka.Event, 1)

	go func(msg *cKafka.Message) {
		defer close(listener)

		e := <-listener
		switch ev := e.(type) {
		case nil:
			return
		case *cKafka.Error:
			log.WithError(ev).
				Error("Processing failed")
			sigchan <- syscall.SIGINT
		case *cKafka.Message:
			if ev.TopicPartition.Error != nil {
				log.WithError(ev.TopicPartition.Error).
					Error("Delivery failed")
				sigchan <- syscall.SIGINT
			} else {
				log.WithFields(log.Fields{
					"key":    string(ev.Key),
					"offset": ev.TopicPartition.Offset,
					"topic":  *ev.TopicPartition.Topic,
				}).
					Debug("Delivered message")
				c.StoreOffset(msg)
			}
		}
	}(msg)

	return listener
}

func (p *Processor) processMessages(producer *FhirProducer, msg *cKafka.Message,
	deliveryChan chan cKafka.Event, sigchan chan os.Signal) {

	bundle := p.mapper.Process(msg.Value)
	if bundle == nil {

		deliveryChan <- nil
		return
	}

	producer.SendBundle(msg.Key, msg.Timestamp, bundle, deliveryChan, sigchan)
}

func syncConsumerCommits(c *ConsentConsumer) {
	c.Unsubscribe()
	parts, err := c.Consumer.Commit()
	if err != nil {
		if err.(cKafka.Error).Code() == cKafka.ErrNoOffset {
			return
		}
		log.WithError(err).Error("Failed to commit offsets")
	} else {

		for _, tp := range parts {
			log.WithFields(log.Fields{
				"topic":     *tp.Topic,
				"partition": tp.Partition,
				"offset":    tp.Offset.String()}).
				Info("Stored offsets committed")
		}
	}
	c.Close()
}
