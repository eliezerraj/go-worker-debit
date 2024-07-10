package main

import(
	"sync"

	"context"
	"time"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/go-worker-debit/internal/util"
	"github.com/go-worker-debit/internal/adapter/event"
	"github.com/go-worker-debit/internal/core"
	"github.com/go-worker-debit/internal/service"
	"github.com/go-worker-debit/internal/repository/postgre"
	"github.com/go-worker-debit/internal/adapter/restapi"
)

var(
	logLevel 	= 	zerolog.DebugLevel
	appServer	core.WorkerAppServer
)

func init() {
	log.Debug().Msg("init")
	zerolog.SetGlobalLevel(logLevel)

	infoPod, restEndpoint := util.GetInfoPod()
	database := util.GetDatabaseEnv()
	configOTEL := util.GetOtelEnv()
	kafkaConfig := util.GetKafkaEnv()

	appServer.InfoPod = &infoPod
	appServer.Database = &database
	appServer.RestEndpoint = &restEndpoint
	appServer.ConfigOTEL = &configOTEL
	appServer.KafkaConfig = &kafkaConfig
}

func main()  {
	log.Debug().Msg("----------------------------------------------------")
	log.Debug().Msg("main")
	log.Debug().Msg("----------------------------------------------------")
	log.Debug().Interface("appServer :",appServer).Msg("")
	log.Debug().Msg("----------------------------------------------------")

	ctx := context.Background()

	// Open Database
	count := 1
	var databaseHelper	postgre.DatabaseHelper
	var err error
	for {
		databaseHelper, err = postgre.NewDatabaseHelper(ctx, appServer.Database)
		if err != nil {
			if count < 3 {
				log.Error().Err(err).Msg("Erro open Database... trying again !!")
			} else {
				log.Error().Err(err).Msg("Fatal erro open Database aborting")
				panic(err)
			}
			time.Sleep(3 * time.Second)
			count = count + 1
			continue
		}
		break
	}

	repoDB := postgre.NewWorkerRepository(databaseHelper)
	restApiService	:= restapi.NewRestApiService()

	workerService := service.NewWorkerService(	&repoDB, 
												appServer.RestEndpoint,
												restApiService)

	consumerWorker, err := event.NewConsumerWorker(ctx, 
													appServer.KafkaConfig, 
													workerService)
	if err != nil {
		log.Error().Err(err).Msg("Erro na abertura do Kafka")
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go consumerWorker.Consumer(	ctx, 
								&wg, 
								appServer)
	wg.Wait()
}