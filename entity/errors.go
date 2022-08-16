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
	ErrUnsupportedFile    = errors.New("UnsupportedFile")
	ErrFileTooLarge       = errors.New("FileTooLarge")
	ErrCoupleNotFound     = errors.New("CoupleNotFound")
	ErrAWSUpload          = errors.New("FileUploadFailed")
	ErrFileProcessing     = errors.New("FailedToProccessFile")
	ErrVideoTooLong       = errors.New("VideoTooLong")
	ErrVideoProcessing    = errors.New("VideoProcessingFailed")
)

type CustomError struct {
	Code    int
	Message string
}

func (c CustomError) Error() string {
	return c.Message
}
