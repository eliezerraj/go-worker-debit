package server

import (
	"fmt"
	"context"
	"sync"
	"encoding/json"

	"github.com/google/uuid"
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

var childLogger = log.With().Str("component","go-worker-debit").Str("package","internal.infra.server").Logger()

var tracerProvider go_core_observ.TracerProvider
var consumerWorker go_core_event.ConsumerWorker
var infoTrace go_core_observ.InfoTrace
var tracer 			trace.Tracer

type ServerWorker struct {
	workerService 	*service.WorkerService
	workerEvent 	*event.WorkerEvent
}

// Set a trace-i inside the context
func setContextTraceId(ctx context.Context, trace_id string) context.Context {
	childLogger.Info().Str("func","setContextTraceId").Interface("trace_id", trace_id).Send()

	var traceUUID string

	if trace_id == "" {
		traceUUID = uuid.New().String()
		trace_id = traceUUID
	}

	ctx = context.WithValue(ctx, "trace-request-id",  trace_id  )
	return ctx
}

// About create a worker event
func NewServerWorker(workerService *service.WorkerService, workerEvent *event.WorkerEvent ) *ServerWorker {
	childLogger.Info().Str("func","NewServerWorker").Send()

	return &ServerWorker{
		workerService: workerService,
		workerEvent: workerEvent,
	}
}

// About consume event kafka
func (s *ServerWorker) Consumer(ctx context.Context, appServer *model.AppServer ,wg *sync.WaitGroup ) {
	childLogger.Info().Str("func","Consumer").Send()

	// otel
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
			childLogger.Info().Err(err).Send()
		}
		childLogger.Info().Msg("closing consumer waiting please !!!")
		defer wg.Done()
	}()

	messages := make(chan go_core_event.Message)

	go s.workerEvent.WorkerKafka.Consumer(s.workerEvent.Topics, messages)

	for msg := range messages {
		childLogger.Info().Msg("=============== MSG FROM KAFKA ==================")
		childLogger.Info().Interface("msg: ",msg).Msg("")
		childLogger.Info().Msg("=============== MSG FROM KAFKA ==================")
		
		// Marshall payload	
		var transfer model.Transfer
		json.Unmarshal([]byte(msg.Payload), &transfer)

		// valid the headers, if there isnt a traceid it will be created
		var header string
		if (*msg.Header)["trace-request-id"] != "" {
			header = (*msg.Header)["trace-request-id"]
		}
		ctx = setContextTraceId(ctx, header)

		//Trace
		ctx, span := tracer.Start(ctx, fmt.Sprintf("go-worker-debit:%v - %v" , transfer.ID, ctx.Value("trace-request-id") ))

		_, err := s.workerService.UpdateDebitMovimentTransfer(ctx, &transfer)
		if err != nil {
			childLogger.Error().Err(err).Interface("trace-resquest-id", ctx.Value("trace-request-id")).Msg("failed update msg: %v : " + msg.Payload)
			childLogger.Info().Interface("trace-resquest-id", ctx.Value("trace-request-id")).Msg("ROLLBACK!!!!")
		} else {
			s.workerEvent.WorkerKafka.Commit()
			childLogger.Info().Interface("trace-resquest-id", ctx.Value("trace-request-id")).Msg("COMMIT!!!!")
		}
		defer span.End()
	}
}