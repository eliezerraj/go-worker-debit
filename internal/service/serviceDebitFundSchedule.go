package service

import (
	"context"
	"github.com/go-worker-debit/internal/core"
	"github.com/go-worker-debit/internal/erro"
	"github.com/aws/aws-xray-sdk-go/xray"

)

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

	_, err = s.restapi.PostData(ctx, s.restapi.ServerUrlDomain ,s.restapi.XApigwId ,"/add", debit)
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
