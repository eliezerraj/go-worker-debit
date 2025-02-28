package server

import (
	"context"
	"sync"
	"encoding/json"

	"github.com/rs/zerolog/log"
	"github.com/go-worker-debit/internal/core/service"
	"github.com/go-worker-debit/internal/adapter/event"
	"github.com/go-worker-debit/internal/core/model"

	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_event "github.com/eliezerraj/go-core/event/kafka" 
)

var childLogger = log.With().Str("infra", "server").Logger()

var tracerProvider go_core_observ.TracerProvider
var consumerWorker go_core_event.ConsumerWorker

type ServerWorker struct {
	workerService 	*service.WorkerService
	workerEvent 	*event.WorkerEvent
}

func NewServerWorker(workerService *service.WorkerService, workerEvent *event.WorkerEvent ) *ServerWorker {
	childLogger.Debug().Msg("NewWorkerEvent")

	return &ServerWorker{
		workerService: workerService,
		workerEvent: workerEvent,
	}
}

func (s *ServerWorker) Consumer(ctx context.Context, wg *sync.WaitGroup ) {
	childLogger.Debug().Msg("Consumer")

	//Trace
	span := tracerProvider.Span(ctx, "service.UpdateCreditMovimentTransfer")
	defer span.End()

	defer func() { 
		span.End()
		childLogger.Debug().Msg("closing consumer waiting please !!!")
		defer wg.Done()
	}()

	messages := make(chan string)

	go s.workerEvent.WorkerKafka.Consumer(s.workerEvent.Topics, messages)

	for msg := range messages {
		childLogger.Debug().Msg("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
		childLogger.Debug().Interface("msg: ",msg).Msg("")
		childLogger.Debug().Msg("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
		
		var transfer model.Transfer
		json.Unmarshal([]byte(msg), &transfer)

		_, err := s.workerService.UpdateDebitMovimentTransfer(ctx, &transfer)
		if err != nil {
			childLogger.Error().Err(err).Msg("failed update msg: %v : " + msg)
			childLogger.Debug().Msg("ROLLBACK!!!!")
		} else {
			s.workerEvent.WorkerKafka.Commit()
			childLogger.Debug().Msg("COMMIT!!!!")
		}
	}
}