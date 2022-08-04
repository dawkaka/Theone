package entity

import "errors"

//ErrNotFound not found
var (
	ErrNotFound           = errors.New("not found")
	ErrPostNotFound       = errors.New("post not foun")
	ErrInvalidEntity      = errors.New("invalid entity")
	ErrCannotBeDeleted    = errors.New("cannot be deleted")
	ErrNotEnoughBooks     = errors.New("not enough books")
	ErrSomethingWentWrong = errors.New("something went wrong")
	ErrNoMatch            = errors.New("no match found")
	ErrInvalidCaption     = errors.New("invalid caption")
	ErrImageProcessing    = errors.New("ImageProcessingFailed")
	ErrInvalidID          = errors.New("invalid bson id")
	ErrUnsupportedImage   = errors.New("UnsupportedImage")
)
