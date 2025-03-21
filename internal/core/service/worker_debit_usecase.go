package service

import(
	"fmt"
	"time"
	"context"
	"net/http"
	"encoding/json"
	"errors"

	"github.com/go-worker-debit/internal/core/model"
	"github.com/go-worker-debit/internal/core/erro"
	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_api "github.com/eliezerraj/go-core/api"
)

var tracerProvider go_core_observ.TracerProvider
var apiService go_core_api.ApiService

func errorStatusCode(statusCode int) error{
	var err error
	switch statusCode {
	case http.StatusUnauthorized:
		err = erro.ErrUnauthorized
	case http.StatusForbidden:
		err = erro.ErrHTTPForbiden
	case http.StatusNotFound:
		err = erro.ErrNotFound
	default:
		err = erro.ErrServer
	}
	return err
}

func (s WorkerService) UpdateDebitMovimentTransfer(ctx context.Context, transfer *model.Transfer) (*model.Transfer, error){
	childLogger.Info().Interface("trace-resquest-id", ctx.Value("trace-request-id")).Msg("UpdateDebitMovimentTransfer")
	childLogger.Info().Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("transfer: ",transfer).Msg("")

	//Trace
	span := tracerProvider.Span(ctx, "service.UpdateDebitMovimentTransfer")
	trace_id := fmt.Sprintf("%v",ctx.Value("trace-request-id"))

	// Get the database connection
	tx, conn, err := s.workerRepository.DatabasePGServer.StartTx(ctx)
	if err != nil {
		return nil, err
	}
	
	// Handle the transaction
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
		s.workerRepository.DatabasePGServer.ReleaseTx(conn)
		span.End()
	}()
	
	// Get transaction UUID 
	res_uuid, err := s.workerRepository.GetTransactionUUID(ctx)
	if err != nil {
		return nil, err
	}

	// Business rule
	time_chargeAt := time.Now()
	transfer.TransactionID = res_uuid
	transfer.Status = "DEBIT_EVENT_DONE"
	transfer.TransferAt = time_chargeAt

	// Get the Account ID from Account-service
	res_acc_from, statusCode, err := apiService.CallApi(ctx,
														s.apiService[0].Url + "/" + transfer.AccountFrom.AccountID,
														s.apiService[0].Method,
														&s.apiService[0].Header_x_apigw_api_id,
														nil,
														&trace_id,
														nil)
	if err != nil {
		return nil, errorStatusCode(statusCode)
	}
	jsonString, err := json.Marshal(res_acc_from)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	var accountStatement model.AccountStatement
	json.Unmarshal(jsonString, &accountStatement)

	transfer.AccountFrom.FkAccountID = accountStatement.ID
	
	// Add (POST) the account statement Get the Account ID from Account-service
	_, statusCode, err = apiService.CallApi(ctx,
											s.apiService[1].Url,
											s.apiService[1].Method,
											&s.apiService[1].Header_x_apigw_api_id,
											nil,
											&trace_id,
											transfer.AccountFrom)
	if err != nil {
		return nil, errorStatusCode(statusCode)
	}

	// Add transfer
	res_transfer, err := s.workerRepository.UpdateDebitMovimentTransfer(ctx, tx, transfer)
	if err != nil {
		return nil, err
	}
	if res_transfer == 0 {
		return nil, erro.ErrUpdate
	}

	return transfer, nil
}