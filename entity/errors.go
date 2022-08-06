package entity

import "errors"

//ErrNotFound not found
var (
	ErrNotFound           = errors.New("NotFound")
	ErrUserNotFound       = errors.New("UserNotFound")
	ErrPostNotFound       = errors.New("PostNotFound")
	ErrSomethingWentWrong = errors.New("SomethingWentWrong")
	ErrNoMatch            = errors.New("NoMatchFound")
	ErrInvalidCaption     = errors.New("InvalidCaption")
	ErrImageProcessing    = errors.New("ImageProcessingFailed")
	ErrInvalidID          = errors.New("InvalidBsonID")
	ErrUnsupportedImage   = errors.New("UnsupportedImage")
)
