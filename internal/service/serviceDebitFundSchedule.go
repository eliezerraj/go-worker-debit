package service

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/go-worker-debit/internal/repository/storage"
	"github.com/go-worker-debit/internal/adapter/restapi"
	"github.com/go-worker-debit/internal/core"
	"github.com/go-worker-debit/internal/erro"
	"github.com/go-worker-debit/internal/lib"
)

var childLogger = log.With().Str("service", "service").Logger()

type WorkerService struct {
	workerRepo		*storage.WorkerRepository
	appServer		*core.WorkerAppServer
	restApiService	*restapi.RestApiService
}

func NewWorkerService(	workerRepo		*storage.WorkerRepository,
						appServer		*core.WorkerAppServer,
						restApiService	*restapi.RestApiService) *WorkerService{
	childLogger.Debug().Msg("NewWorkerService")

	return &WorkerService{
		workerRepo:	workerRepo,
		appServer:	appServer,
		restApiService:		restApiService,
	}
}

func (s WorkerService) DebitFundSchedule(ctx context.Context, transfer core.Transfer) (error){
	childLogger.Debug().Msg("DebitFundSchedule")
	childLogger.Debug().Interface("===>transfer:",transfer).Msg("")
	
	span := lib.Span(ctx, "service.DebitFundSchedule")

	tx, conn, err := s.workerRepo.StartTx(ctx)
	if err != nil {
		return err
	}
	
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
		s.workerRepo.ReleaseTx(conn)
		span.End()
	}()

	// Post
	debit := core.AccountStatement{}
	debit.AccountID = transfer.AccountIDTo
	debit.Currency = transfer.Currency
	debit.Amount = transfer.Amount
	debit.Type = transfer.Type
	transfer.Status = "DEBIT_DONE"

	path := s.appServer.RestEndpoint.ServiceUrlDomain + "/add"
	_, err = s.restApiService.CallRestApi(ctx,"POST",path, &s.appServer.RestEndpoint.XApigwId,debit)
	if err != nil {
		switch err{
			case erro.ErrTransInvalid:
				transfer.Status = "DEBIT_FAIL_MISMATCH_DATA"
			default:
				transfer.Status = "DEBIT_FAIL_OUTAGE"
			}
	}

	res_update, err := s.workerRepo.Update(ctx, tx, &transfer)
	if err != nil {
		return err
	}
	if res_update == 0 {
		err = erro.ErrUpdate
		return err
	}

	if transfer.Status != "DEBIT_DONE"{
		return erro.ErrEvent
	}

	return nil
}