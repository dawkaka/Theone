package entity

import "errors"

//ErrNotFound not found
var ErrNotFound = errors.New("not found")
var ErrPostNotFound = errors.New("post not foun")

//ErrInvalidEntity invalid entity
var ErrInvalidEntity = errors.New("invalid entity")

//ErrCannotBeDeleted cannot be deleted
var ErrCannotBeDeleted = errors.New("cannot be deleted")

//ErrNotEnoughBooks cannot borrow
var ErrNotEnoughBooks = errors.New("not enough books")

//Something Went Wrong Error
var ErrSomethingWentWrong = errors.New("something went wrong")
