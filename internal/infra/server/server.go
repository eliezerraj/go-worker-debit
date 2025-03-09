package server

import (
	"fmt"
	"context"
	"sync"
	"encoding/json"

	"github.com/rs/zerolog/log"
	"github.com/go-worker-debit/internal/core/service"
	"github.com/go-worker-debit/internal/adapter/event"
	"github.com/go-worker-debit/internal/core/model"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/contrib/propagators/aws/xray"

	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_event "github.com/eliezerraj/go-core/event/kafka" 
)

var childLogger = log.With().Str("infra", "server").Logger()

var tracerProvider go_core_observ.TracerProvider
var consumerWorker go_core_event.ConsumerWorker
var infoTrace go_core_observ.InfoTrace
var tracer 			trace.Tracer

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

func (s *ServerWorker) Consumer(ctx context.Context, appServer *model.AppServer ,wg *sync.WaitGroup ) {
	childLogger.Debug().Msg("Consumer")

	// otel
	childLogger.Info().Str("OTEL_EXPORTER_OTLP_ENDPOINT :", appServer.ConfigOTEL.OtelExportEndpoint).Msg("")

	infoTrace.PodName = appServer.InfoPod.PodName
	infoTrace.PodVersion = appServer.InfoPod.ApiVersion
	infoTrace.ServiceType = "k8-workload"
	infoTrace.Env = appServer.InfoPod.Env
	infoTrace.AccountID = appServer.InfoPod.AccountID

	tp := tracerProvider.NewTracerProvider(	ctx, 
											appServer.ConfigOTEL, 
											&infoTrace)
	otel.SetTextMapPropagator(xray.Propagator{})
	otel.SetTracerProvider(tp)
	tracer = tp.Tracer(appServer.InfoPod.PodName)

	// handle defer
	defer func() { 
		err := tp.Shutdown(ctx)
		if err != nil{
			childLogger.Error().Err(err).Msg("error closing OTEL tracer !!!")
		}
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

			//Trace
		ctx, span := tracer.Start(ctx, fmt.Sprintf("go-worker-debit:%v" , transfer.TransactionID ))
		
		_, err := s.workerService.UpdateDebitMovimentTransfer(ctx, &transfer)
		if err != nil {
			childLogger.Error().Err(err).Msg("failed update msg: %v : " + msg)
			childLogger.Debug().Msg("ROLLBACK!!!!")
		} else {
			s.workerEvent.WorkerKafka.Commit()
			childLogger.Debug().Msg("COMMIT!!!!")
		}
		defer span.End()
	}
}