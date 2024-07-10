package event

import (
	"os"
	"os/signal"
	"syscall"
	"sync"
	"context"
	"encoding/json"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-worker-debit/internal/core"
	"github.com/go-worker-debit/internal/service"
	"github.com/go-worker-debit/internal/lib"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
)

var childLogger = log.With().Str("adpater", "kafka").Logger()

type ConsumerWorker struct{
	configurations  *core.KafkaConfig
	consumer        *kafka.Consumer
	workerService	*service.WorkerService
	tracer 			trace.Tracer
}

func NewConsumerWorker(	ctx context.Context, 
						configurations *core.KafkaConfig,
						workerService	*service.WorkerService ) (*ConsumerWorker, error) {
	childLogger.Debug().Msg("NewConsumerWorker")

	kafkaBrokerUrls := 	configurations.KafkaConfigurations.Brokers1 + "," + configurations.KafkaConfigurations.Brokers2 + "," + configurations.KafkaConfigurations.Brokers3
	config := &kafka.ConfigMap{	"bootstrap.servers":            kafkaBrokerUrls,
								"security.protocol":            configurations.KafkaConfigurations.Protocol, //"SASL_SSL",
								"sasl.mechanisms":              configurations.KafkaConfigurations.Mechanisms, //"SCRAM-SHA-256",
								"sasl.username":                configurations.KafkaConfigurations.Username,
								"sasl.password":                configurations.KafkaConfigurations.Password,
								"group.id":                     configurations.KafkaConfigurations.Groupid,
								"enable.auto.commit":           false, //true,
								"broker.address.family": 		"v4",
								"client.id": 					configurations.KafkaConfigurations.Clientid,
								"session.timeout.ms":    		6000,
								"enable.idempotence":			true,
								"auto.offset.reset":     		"earliest", //"latest",  
								}

	consumer, err := kafka.NewConsumer(config)
	if err != nil {
		childLogger.Error().Err(err).Msg("Failed to create consumer")
		return nil, err
	}

	return &ConsumerWorker{ configurations: configurations,
							consumer: 		consumer,
							workerService: 	workerService,
	}, nil
}

func (c *ConsumerWorker) Consumer(ctx context.Context, wg *sync.WaitGroup, appServer core.WorkerAppServer) {
	childLogger.Debug().Msg("Consumer")

	// ---------------------- OTEL ---------------
	childLogger.Info().Str("OTEL_EXPORTER_OTLP_ENDPOINT :", appServer.ConfigOTEL.OtelExportEndpoint).Msg("")
	
	tp := lib.NewTracerProvider(ctx, appServer.ConfigOTEL, appServer.InfoPod)
	defer func() { 
		err := tp.Shutdown(ctx)
		if err != nil{
			childLogger.Error().Err(err).Msg("Erro closing OTEL tracer !!!")
		}
	}()
	otel.SetTextMapPropagator(xray.Propagator{})
	otel.SetTracerProvider(tp)

	topics := []string{appServer.KafkaConfig.Topic.Debit}
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	err := c.consumer.SubscribeTopics(topics, nil)
	if err != nil {
		childLogger.Error().Err(err).Msg("Failed to subscriber topic")
	}

	run := true
	for run {
		select {
			case sig := <-sigchan:
				childLogger.Debug().Interface("Caught signal terminating: ", sig).Msg("")
				run = false
			default:
				ev := c.consumer.Poll(100)
				if ev == nil {
					continue
				}
			switch e := ev.(type) {
				case kafka.AssignedPartitions:
					c.consumer.Assign(e.Partitions)
				case kafka.RevokedPartitions:
					c.consumer.Unassign()	
				case kafka.PartitionEOF:
					childLogger.Error().Interface("kafka.PartitionEOF: ",e).Msg("")
				case *kafka.Message:
					
					log.Print("----------------------------------")
					if e.Headers != nil {
						log.Printf("Headers: %v\n", e.Headers)
					}
					log.Print("Value : " ,string(e.Value))
					log.Print("-----------------------------------")
					
					event := core.Event{}
					json.Unmarshal(e.Value, &event)

					tracer := tp.Tracer("go-worker-debit:" + event.EventData.Transfer.AccountIDTo + ":" + strconv.Itoa(event.EventData.Transfer.ID))
					ctx, span := tracer.Start(ctx, "go-worker-debit")
					defer span.End()

					err = c.workerService.DebitFundSchedule(ctx, *event.EventData.Transfer)
					if err != nil {
						childLogger.Error().Err(err).Msg("Erro no service.DebitFundSchedule ROLLBACK !!!!")
						childLogger.Debug().Msg("ROLLBACK!!!!")	
					} else {
						childLogger.Debug().Msg("COMMIT!!!!")
						c.consumer.Commit()
					}

				case kafka.Error:
					childLogger.Error().Err(e).Msg("kafka.Error")
					if e.Code() == kafka.ErrAllBrokersDown {
						run = false
					}
				default:
					childLogger.Debug().Interface("default: ",e).Msg("Ignored")
			}
		}
	}

	childLogger.Debug().Msg("Closing consumer waiting please !!!")
	c.consumer.Close()
	defer wg.Done()
}