package jsonrpc

type ErrorMessage string

const (
	INSUFFICIENT_PRIVILEGES ErrorMessage = "Insufficient privileges. Please contact the owner of your company account"
	NOT_VERIFIED            ErrorMessage = "Please verify your email: "
	ALREADY_VERIFIED        ErrorMessage = "The email associated with this account has already been verified"
	UNASSIGNED              ErrorMessage = "This account is not assigned to a company. Navigate to the settings tab to complete company assignment"
	COMPANY_ALREADY_EXISTS  ErrorMessage = "This company code is already in use. Please try again using a different code"
	INVALID_EMPTY           ErrorMessage = "A value is required to complete this request. Please try again with a valid entry"
	INVALID_TOKEN           ErrorMessage = "There is something wrong with your session token. Please sign out and log back in again"
	TOKEN_UPDATE            ErrorMessage = "Something went wrong updating this session's token. Please try the last operation again or contact us through the messenger in the bottom right corner of the screen"
	INVALID_DOMAIN          ErrorMessage = "This account's email domain is not listed as valid for the following company code: "
	AUTH0_FAILURE           ErrorMessage = "Something went wrong. Please try again later or contact us through the messenger in the bottom right corner of the screen"
	FAILURE                 ErrorMessage = "Something went wrong. Please try again later or contact us through the messenger in the bottom right corner of the screen"
)

func (e ErrorMessage) ToString() string {
	return string(e)
}
