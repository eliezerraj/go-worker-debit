package erro

import (
	"errors"

)

var (
	ErrHTTPForbiden		= errors.New("Requisição recusada")
	ErrServer		 	= errors.New("Erro não identificado")
	ErrNotFound 		= errors.New("Item não encontrado")
	ErrUnauthorized 	= errors.New("Erro de autorização")
	ErrDecode			= errors.New("Erro na decodificação do Base64")
	ErrUpdate			= errors.New("Erro no update do dado")
	ErrTransInvalid		= errors.New("Erro de transação invalida")
	ErrEvent			= errors.New("Erro no tratamento de evento")
)
