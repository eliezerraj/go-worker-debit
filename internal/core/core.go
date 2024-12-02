package core

import (
	"time"
)

type DatabaseRDS struct {
    Host 				string `json:"host"`
    Port  				string `json:"port"`
	Schema				string `json:"schema"`
	DatabaseName		string `json:"databaseName"`
	User				string `json:"user"`
	Password			string `json:"password"`
	Db_timeout			int	`json:"db_timeout"`
	Postgres_Driver		string `json:"postgres_driver"`
}

type WorkerAppServer struct {
	InfoPod 		*InfoPod 		`json:"info_pod"`
	Database		*DatabaseRDS	`json:"database"`
	RestEndpoint	*RestEndpoint	`json:"rest_endpoint"`
	ConfigOTEL		*ConfigOTEL		`json:"otel_config"`
	KafkaConfig		*KafkaConfig	`json:"kafka_config"`
	AwsServiceConfig *AwsServiceConfig	`json:"aws_service_config"`
	RestApiCallData *RestApiCallData `json:"rest_api_call_dsa_data"`
}

type InfoPod struct {
	PodName				string `json:"pod_name"`
	ApiVersion			string `json:"version"`
	OSPID				string `json:"os_pid"`
	IPAddress			string `json:"ip_address"`
	AvailabilityZone 	string `json:"availabilityZone"`
	IsAZ				bool   	`json:"is_az"`
	Env					string `json:"enviroment,omitempty"`
	AccountID			string `json:"account_id,omitempty"`
}

type RestEndpoint struct {
	ServiceUrlDomain 		string `json:"service_url_domain"`
	XApigwId				string `json:"xApigwId"`
	ServerHost			string `json:"server_host_localhost,omitempty"`
}

type KafkaConfig struct {
	KafkaConfigurations    	*KafkaConfigurations
	Topic					*Topic `json:"topic"`
}

type KafkaConfigurations struct {
    Username		string 
    Password		string 
    Protocol		string
    Mechanisms		string
    Clientid		string 
    Brokers1		string 
    Brokers2		string 
    Brokers3		string 
	Groupid			string 
	Partition       int
    ReplicationFactor int
    RequiredAcks    int
    Lag             int
    LagCommit       int
}

type Event struct {
    ID          int         `json:"id"`
	Key			string      `json:"key"`
    EventDate   time.Time   `json:"event_date"`
    EventType   string      `json:"event_type"`
    EventData   *EventData   `json:"event_data"`
}

type EventData struct {
    Transfer   *Transfer    `json:"transfer"`
}

type Topic struct {
	Debit     string    `json:"topic_debit"`
}

type ConfigOTEL struct {
	OtelExportEndpoint		string
	TimeInterval            int64    `mapstructure:"TimeInterval"`
	TimeAliveIncrementer    int64    `mapstructure:"RandomTimeAliveIncrementer"`
	TotalHeapSizeUpperBound int64    `mapstructure:"RandomTotalHeapSizeUpperBound"`
	ThreadsActiveUpperBound int64    `mapstructure:"RandomThreadsActiveUpperBound"`
	CpuUsageUpperBound      int64    `mapstructure:"RandomCpuUsageUpperBound"`
	SampleAppPorts          []string `mapstructure:"SampleAppPorts"`
}

type Transfer struct {
	ID				int			`json:"id,omitempty"`
	AccountIDFrom	string		`json:"account_id_from,omitempty"`
	FkAccountIDFrom	int			`json:"fk_account_id_from,omitempty"`
	TransferAt		time.Time 	`json:"transfer_at,omitempty"`
	Type			string  	`json:"type_charge,omitempty"`
	Status			string  	`json:"status,omitempty"`
	Currency		string  	`json:"currency,omitempty"`
	Amount			float64 	`json:"amount,omitempty"`
	AccountIDTo		string		`json:"account_id_to,omitempty"`
	FkAccountIDTo	int			`json:"fk_account_id_to,omitempty"`
}

type AccountStatement struct {
	ID				int			`json:"id,omitempty"`
	FkAccountID		int			`json:"fk_account_id,omitempty"`
	AccountID		string		`json:"account_id,omitempty"`
	Type			string  	`json:"type_charge,omitempty"`
	ChargeAt		time.Time 	`json:"charged_at,omitempty"`
	Currency		string  	`json:"currency,omitempty"`
	Amount			float64 	`json:"amount,omitempty"`
	TenantID		string  	`json:"tenant_id,omitempty"`
}

type AwsServiceConfig struct {
	AwsRegion				string	`json:"aws_region"`
	ServiceUrlJwtSA 		string	`json:"service_url_jwt_sa"`
	SecretJwtSACredential 	string	`json:"secret_jwt_credential"`
	UsernameJwtDA			string	`json:"username_jwt_sa"`
	PasswordJwtDA			string	`json:"password_jwt_sa"`
}

type TokenSA struct {
	Token string `json:"token,omitempty"`
	Err   error
}

type RestApiCallData struct {
	Url				string `json:"url"`
	Method			string `json:"method"`
	X_Api_Id		*string `json:"x-apigw-api-id"`
	UsernameAuth	string `json:"user"`
	PasswordAuth 	string `json:"password"`
}