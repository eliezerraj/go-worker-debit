package service

import (
	"github.com/rs/zerolog/log"
	"github.com/go-worker-debit/internal/repository/postgre"
	"github.com/go-worker-debit/internal/adapter/restapi"

)

var childLogger = log.With().Str("service", "service").Logger()

type WorkerService struct {
	workerRepository 		*db_postgre.WorkerRepository
	restapi					*restapi.RestApiSConfig
}

func NewWorkerService(workerRepository *db_postgre.WorkerRepository,
					restapi				*restapi.RestApiSConfig ) *WorkerService{
	childLogger.Debug().Msg("NewWorkerService")

	return &WorkerService{
		workerRepository:	workerRepository,
		restapi:			restapi,
	}
}
