package main

import(
	"time"
	"context"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/go-worker-debit/internal/infra/configuration"
	"github.com/go-worker-debit/internal/core/model"
	"github.com/go-worker-debit/internal/core/service"
	"github.com/go-worker-debit/internal/adapter/database"
	"github.com/go-worker-debit/internal/adapter/event"
	"github.com/go-worker-debit/internal/infra/server"
	go_core_pg "github.com/eliezerraj/go-core/database/pg"  
)

var(
	logLevel = 	zerolog.DebugLevel
	appServer	model.AppServer
	databaseConfig go_core_pg.DatabaseConfig
	databasePGServer go_core_pg.DatabasePGServer
)

func init(){
	log.Debug().Msg("init")
	zerolog.SetGlobalLevel(logLevel)

	infoPod := configuration.GetInfoPod()
	configOTEL 		:= configuration.GetOtelEnv()
	databaseConfig 	:= configuration.GetDatabaseEnv() 
	apiService 	:= configuration.GetEndpointEnv() 
	kafkaConfigurations, topics := configuration.GetKafkaEnv() 

	appServer.InfoPod = &infoPod
	appServer.ConfigOTEL = &configOTEL
	appServer.DatabaseConfig = &databaseConfig
	appServer.ApiService = apiService
	appServer.KafkaConfigurations = &kafkaConfigurations
	appServer.Topics = topics
}

func main (){
	log.Debug().Msg("----------------------------------------------------")
	log.Debug().Msg("main")
	log.Debug().Msg("----------------------------------------------------")
	log.Debug().Interface("appServer :",appServer).Msg("")
	log.Debug().Msg("----------------------------------------------------")

	ctx := context.Background()

	// Open Database
	count := 1
	var err error
	for {
		databasePGServer, err = databasePGServer.NewDatabasePGServer(ctx, *appServer.DatabaseConfig)
		if err != nil {
			if count < 3 {
				log.Error().Err(err).Msg("error open database... trying again !!")
			} else {
				log.Error().Err(err).Msg("fatal error open Database aborting")
				panic(err)
			}
			time.Sleep(3 * time.Second) //backoff
			count = count + 1
			continue
		}
		break
	}

	// Database
	database := database.NewWorkerRepository(&databasePGServer)
	workerService := service.NewWorkerService(database, appServer.ApiService)

	// Kafka
	workerEvent, err := event.NewWorkerEvent(ctx, appServer.Topics, appServer.KafkaConfigurations)
	if err != nil {
		log.Error().Err(err).Msg("error open kafka")
		panic(err)
	}

	serverWorker := server.NewServerWorker(workerService, workerEvent)

	var wg sync.WaitGroup
	wg.Add(1)
	go serverWorker.Consumer(ctx, &wg)
	wg.Wait()
}