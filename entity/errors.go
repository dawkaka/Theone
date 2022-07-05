package entity

import "errors"

//ErrNotFound not found
var ErrNotFound = errors.New("not found")

//ErrInvalidEntity invalid entity
var ErrInvalidEntity = errors.New("invalid entity")

//ErrCannotBeDeleted cannot be deleted
var ErrCannotBeDeleted = errors.New("cannot be deleted")

//ErrNotEnoughBooks cannot borrow
var ErrNotEnoughBooks = errors.New("not enough books")

//ErrBookAlreadyBorrowed cannot borrow
var ErrBookAlreadyBorrowed = errors.New("book already borrowed")

//ErrBookNotBorrowed cannot return
var ErrBookNotBorrowed = errors.New("book not borrowed")
