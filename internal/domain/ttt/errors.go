package ttt

import "errors"

var (


	ErrInvalidMove  = errors.New("invalid move")
	ErrCellOccupied = errors.New("cell is already occupied")
	ErrOutOfBounds  = errors.New("coordinates out of bounds")
)
