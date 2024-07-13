package restapi

import(
	"errors"
	"net/http"
	"time"
	"encoding/json"
	"bytes"
	"context"

	"github.com/rs/zerolog/log"
	"github.com/go-worker-debit/internal/erro"
	"github.com/go-worker-debit/internal/lib"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var childLogger = log.With().Str("adapter/restapi", "restapi").Logger()
//--------------------------------------------------
type RestApiService struct {

}

func NewRestApiService() (*RestApiService){
	childLogger.Debug().Msg("*** NewRestApiService")
	
	return &RestApiService {
	}
}
//--------------------------------------------------
func (r *RestApiService) GetData(ctx context.Context, urlDomain string, xApigwId string, data interface{}) (interface{}, error) {
	childLogger.Debug().Msg("GetData")

	data_interface, err := makeGet(ctx, urlDomain, xApigwId ,data)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Request")
		return nil, err
	}
    
	return data_interface, nil
}

func (r *RestApiService) PostData(ctx context.Context, urlDomain string, xApigwId string, data interface{}) (interface{}, error) {
	childLogger.Debug().Msg("PostData")

	data_interface, err := makePost(ctx, urlDomain, xApigwId ,data)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Request")
		return nil, err
	}
    
	return data_interface, nil
}

func makeGet(ctx context.Context, url string, xApigwId string, data interface{}) (interface{}, error) {
	childLogger.Debug().Msg("makeGet")
	childLogger.Debug().Str("url : ", url).Msg("")
	childLogger.Debug().Str("xApigwId : ", xApigwId).Msg("")

	span := lib.Span(ctx, url)	
    defer span.End()

	client := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Request")
		return false, errors.New(err.Error())
	}
	req.Header.Add("Content-Type", "application/json;charset=UTF-8");
	req.Header.Add("x-apigw-api-id", xApigwId);

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		childLogger.Error().Err(err).Msg("error Do Request")
		return false, errors.New(err.Error())
	}
	defer resp.Body.Close()

	childLogger.Debug().Int("StatusCode :", resp.StatusCode).Msg("")
	switch (resp.StatusCode) {
		case 401:
			return false, erro.ErrHTTPForbiden
		case 403:
			return false, erro.ErrHTTPForbiden
		case 200:
		case 400:
			return false, erro.ErrNotFound
		case 404:
			return false, erro.ErrNotFound
		case 409:
			return false, erro.ErrTransInvalid
		default:
			return false, erro.ErrServer
	}

	result := data
	err = json.NewDecoder(resp.Body).Decode(&result)
    if err != nil {
		childLogger.Error().Err(err).Msg("error no ErrUnmarshal")
		return false, errors.New(err.Error())
    }

	return result, nil
}

func makePost(ctx context.Context, url string, xApigwId string, data interface{}) (interface{}, error) {
	childLogger.Debug().Msg("makePost")
	childLogger.Debug().Str("url : ", url).Msg("")
	childLogger.Debug().Str("xApigwId : ", xApigwId).Msg("")

	span := lib.Span(ctx, url)	
    defer span.End()

	client := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout: time.Second * 10,
	}

	payload := new(bytes.Buffer)
	json.NewEncoder(payload).Encode(data)

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Request")
		return false, errors.New(err.Error())
	}
	req.Header.Add("Content-Type", "application/json;charset=UTF-8");
	req.Header.Add("x-apigw-api-id", xApigwId);

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		childLogger.Error().Err(err).Msg("error Do Request")
		return false, errors.New(err.Error())
	}
	defer resp.Body.Close()

	childLogger.Debug().Int("StatusCode :", resp.StatusCode).Msg("")
	switch (resp.StatusCode) {
		case 401:
			return false, erro.ErrHTTPForbiden
		case 403:
			return false, erro.ErrHTTPForbiden
		case 200:
		case 400:
			return false, erro.ErrNotFound
		case 404:
			return false, erro.ErrNotFound
		case 409:
			return false, erro.ErrTransInvalid
		default:
			return false, erro.ErrHTTPForbiden
	}

	result := data
	err = json.NewDecoder(resp.Body).Decode(&result)
    if err != nil {
		childLogger.Error().Err(err).Msg("error no ErrUnmarshal")
		return false, errors.New(err.Error())
    }

	return result, nil
}
