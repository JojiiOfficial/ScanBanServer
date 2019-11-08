package main

//ErrorMessage err message
type ErrorMessage string

//ResponseError response for errors
var ResponseError = Status{"error", "An Servererror occurred"}

//ResponseSuccess response for success but no data
var ResponseSuccess = Status{"success", ""}

const (
	//ServerError error from server
	ServerError ErrorMessage = "Server Error"
	//WrongInputFormatError wrong user input
	WrongInputFormatError ErrorMessage = "Wrong inputFormat!"
	//NoValidIPFound no valid ip found in report
	NoValidIPFound ErrorMessage = "No valid Ip found!"
	//InvalidTokenError token is not valid
	InvalidTokenError ErrorMessage = "Token not valid"
	//BatchSizeTooLarge batch is too large
	BatchSizeTooLarge ErrorMessage = "BatchSize soo large!"
	//WrongIntegerFormat integer is probably no integer
	WrongIntegerFormat ErrorMessage = "Number is string"
)
