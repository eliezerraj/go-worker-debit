package service

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/go-worker-debit/internal/repository/postgre"
	"github.com/go-worker-debit/internal/adapter/restapi"
	"github.com/go-worker-debit/internal/core"
	"github.com/go-worker-debit/internal/erro"
	"github.com/aws/aws-xray-sdk-go/xray"

)

var childLogger = log.With().Str("service", "service").Logger()

type WorkerService struct {
	workerRepository 		*postgre.WorkerRepository
	restEndpoint			*core.RestEndpoint
	restApiService			*restapi.RestApiService
}

func NewWorkerService(	workerRepository *postgre.WorkerRepository,
						restEndpoint		*core.RestEndpoint,
						restApiService		*restapi.RestApiService) *WorkerService{
	childLogger.Debug().Msg("NewWorkerService")

	return &WorkerService{
		workerRepository:	workerRepository,
		restEndpoint:		restEndpoint,
		restApiService:		restApiService,
	}
}

func (s WorkerService) DebitFundSchedule(ctx context.Context, transfer core.Transfer) (error){
	childLogger.Debug().Msg("DebitFundSchedule")

	childLogger.Debug().Interface("===>transfer:",transfer).Msg("")
	
	_, root := xray.BeginSubsegment(ctx, "Service.DebitFundSchedule")
	defer root.Close(nil)

	tx, err := s.workerRepository.StartTx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// Post
	debit := core.AccountStatement{}
	debit.AccountID = transfer.AccountIDTo
	debit.Currency = transfer.Currency
	debit.Amount = transfer.Amount
	debit.Type = transfer.Type
	transfer.Status = "DEBIT_DONE"

	_, err = s.restApiService.PostData(ctx, 
										s.restEndpoint.ServiceUrlDomain,
										s.restEndpoint.XApigwId,
										"/add", 
										debit)
	if err != nil {
		switch err{
			case erro.ErrTransInvalid:
				transfer.Status = "DEBIT_FAIL_MISMATCH_DATA"
			default:
				transfer.Status = "DEBIT_FAIL_OUTAGE"
			}
	}

	res_update, err := s.workerRepository.Update(ctx,tx ,transfer)
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
