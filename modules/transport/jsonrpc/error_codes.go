package jsonrpc

import (
	"github.com/gorilla/rpc/v2/json2"
)

type JSONRPCErrorData struct {
	Name         string            `json:"name"`
	MissingField string            `json:"missing_fields"`
	Meta         map[string]string `json:"meta"`
}

const (
	ERROR_UNKNOWN                     json2.ErrorCode = 0
	ERROR_INSUFFICIENT_PRIVILEGES     json2.ErrorCode = 1
	ERROR_AUTH0_FAILURE               json2.ErrorCode = 2
	ERROR_JWT_PARSE_FAILURE           json2.ErrorCode = 3
	ERROR_USER_IS_NOT_ASSIGNED        json2.ErrorCode = 4
	ERROR_STORAGE_FAILURE             json2.ErrorCode = 5
	ERROR_MISSING_FIELD               json2.ErrorCode = 6
	ERROR_PASSWORD_GENERATION_FAILURE json2.ErrorCode = 7
	ERROR_ILLEGAL_OPERATION           json2.ErrorCode = 8
	ERROR_SLACK_FAILURE               json2.ErrorCode = 9
	ERROR_HUBSPOT_FAILURE             json2.ErrorCode = 10
)

var JSONRPCErrorCodes []json2.Error = []json2.Error{
	{
		Message: "Unknown error",
		Code:    ERROR_UNKNOWN,
		Data: &JSONRPCErrorData{
			Name: "ERROR_UNKNOWN",
		},
	},
	{
		Message: "The user account does not have sufficient privileges to make this request",
		Code:    ERROR_INSUFFICIENT_PRIVILEGES,
		Data: &JSONRPCErrorData{
			Name: "ERROR_INSUFFICIENT_PRIVILEGES",
		},
	},
	{
		Message: "A call to Auth0 failed",
		Code:    ERROR_AUTH0_FAILURE,
		Data: &JSONRPCErrorData{
			Name: "ERROR_AUTH0_FAILURE",
		},
	},
	{
		Message: "Parsing the request JWT failed",
		Code:    ERROR_JWT_PARSE_FAILURE,
		Data: &JSONRPCErrorData{
			Name: "ERROR_JWT_PARSE_FAILURE",
		},
	},
	{
		Message: "The user account is not assigned to a company",
		Code:    ERROR_USER_IS_NOT_ASSIGNED,
		Data: &JSONRPCErrorData{
			Name: "ERROR_USER_IS_NOT_ASSIGNED",
		},
	},
	{
		Message: "A storage request has failed",
		Code:    ERROR_STORAGE_FAILURE,
		Data: &JSONRPCErrorData{
			Name: "ERROR_STORAGE_FAILURE",
		},
	},
	{
		Message: "The request is missing a required field",
		Code:    ERROR_MISSING_FIELD,
		Data: &JSONRPCErrorData{
			Name: "ERROR_MISSING_FIELD",
		},
	},
	{
		Message: "Failed to generate a random string",
		Code:    ERROR_PASSWORD_GENERATION_FAILURE,
		Data: &JSONRPCErrorData{
			Name: "ERROR_PASSWORD_GENERATION_FAILURE",
		},
	},
	{
		Message: "This operation is not supported",
		Code:    ERROR_ILLEGAL_OPERATION,
		Data: &JSONRPCErrorData{
			Name: "ERROR_ILLEGAL_OPERATION",
		},
	},
	{
		Message: "Slack notification failed to send",
		Code:    ERROR_SLACK_FAILURE,
		Data: &JSONRPCErrorData{
			Name: "ERROR_SLACK_FAILURE",
		},
	},
}
