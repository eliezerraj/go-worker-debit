package configuration

import(
	"os"
	"github.com/joho/godotenv"
	"github.com/go-worker-debit/internal/core/model"
)

func GetEndpointEnv() []model.ApiService {
	childLogger.Info().Str("func","GetEndpointEnv").Send()

	err := godotenv.Load(".env")
	if err != nil {
		childLogger.Info().Err(err).Send()
	}
	
	var apiService []model.ApiService

	var apiService01 model.ApiService
	if os.Getenv("URL_SERVICE_01") !=  "" {
		apiService01.Url = os.Getenv("URL_SERVICE_01")
	}
	if os.Getenv("X_APIGW_API_ID_SERVICE_01") !=  "" {
		apiService01.Header_x_apigw_api_id = os.Getenv("X_APIGW_API_ID_SERVICE_01")
	}
	if os.Getenv("METHOD_SERVICE_01") !=  "" {
		apiService01.Method = os.Getenv("METHOD_SERVICE_01")
	}
	if os.Getenv("NAME_SERVICE_01") !=  "" {
		apiService01.Name = os.Getenv("NAME_SERVICE_01")
	}
	apiService = append(apiService, apiService01)

	var apiService02 model.ApiService
	if os.Getenv("URL_SERVICE_02") !=  "" {
		apiService02.Url = os.Getenv("URL_SERVICE_02")
	}
	if os.Getenv("X_APIGW_API_ID_SERVICE_02") !=  "" {
		apiService02.Header_x_apigw_api_id = os.Getenv("X_APIGW_API_ID_SERVICE_02")
	}
	if os.Getenv("METHOD_SERVICE_02") !=  "" {
		apiService02.Method = os.Getenv("METHOD_SERVICE_02")
	}
	if os.Getenv("NAME_SERVICE_02") !=  "" {
		apiService02.Name = os.Getenv("NAME_SERVICE_02")
	}
	apiService = append(apiService, apiService02)
	
	return apiService
}