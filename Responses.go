package main

//ErrorMessage err message
type ErrorMessage string

//ResponseError response for errors
var ResponseError = Status{"error", "An Servererror occurred"}

//ResponseSuccess response for success but no data
var ResponseSuccess = Status{"success", ""}

const (
	ServerError           ErrorMessage = "Server Error"
	WrongInputFormatError ErrorMessage = "Wrong inputFormat!"
	EmptyError            ErrorMessage = "Wrong inputFormat!"
	InvalidTokenError     ErrorMessage = "Token not valid"
	PlaceNotFond          ErrorMessage = "Place not found"
	PlaceAlreadyExists    ErrorMessage = "Place already exists"
	BatchSizeTooLarge     ErrorMessage = "BatchSize soo large!"
	WrongIntegerFormat    ErrorMessage = "Number is string"
)
