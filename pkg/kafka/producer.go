package kafka

import (
	"consent-to-fhir/pkg/config"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

type FhirProducer struct {
	Producer *kafka.Producer
	Topic    string
}

func NewProducer(config config.Kafka) *FhirProducer {

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":        config.BootstrapServers,
		"security.protocol":        config.SecurityProtocol,
		"ssl.ca.location":          config.Ssl.CaLocation,
		"ssl.key.location":         config.Ssl.KeyLocation,
		"ssl.certificate.location": config.Ssl.CertificateLocation,
		"ssl.key.password":         config.Ssl.KeyPassword,
	})
	if err != nil {
		log.WithError(err).Error("Failed to create Kafka producer. Terminating")
		os.Exit(1)
	}

	return &FhirProducer{
		Producer: p,
		Topic:    config.OutputTopic,
	}
}

func (p *FhirProducer) SendBundle(key []byte, timestamp time.Time, bundle *fhir.Bundle, deliveryChan chan kafka.Event, sigchan chan os.Signal) {
	if bundle != nil {
		byteVal, err := bundle.MarshalJSON()
		if err != nil {
			log.WithError(err).Error("Failed to serialize Bundle to JSON")
			// TODO
			deliveryChan <- kafka.Error{}
			return
		}
		p.Send(key, timestamp, byteVal, deliveryChan, sigchan)
	}
}

func (p *FhirProducer) Send(key []byte, timestamp time.Time, msg []byte, deliveryChan chan kafka.Event, sigchan chan os.Signal) {

	err := p.Producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &p.Topic, Partition: kafka.PartitionAny},
		Key:            key,
		Timestamp:      timestamp,
		Value:          msg,
	}, deliveryChan)
	if err != nil {
		if err.(kafka.Error).Code() == kafka.ErrQueueFull {
			// Producer queue is full, wait 1s for messages
			// to be delivered then try again.
			time.Sleep(time.Second)
			p.Send(key, timestamp, msg, deliveryChan, sigchan)
		}
	}

	select {
	case <-sigchan:
		return
	case <-deliveryChan:
		return
	}
}
