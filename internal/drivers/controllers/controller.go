package controllers

import (
	"io"
)

// Definitions for IOControl constants and function taken from linux kernel:
// https://github.com/torvalds/linux/blob/master/include/uapi/asm-generic/ioctl.h
//
// Note that I changed the name "type" to "magic" because type is a reserved
// word in Go. The "type" field is described as a "magic" number specific to
// a driver or subsystem used to ensure uniqueness and prevent the command
// from going to an unintended device
const (
	IOControlNumberBits = uint(8)
	IOControlMagicBits  = uint(8)
	IOControlSizeBits   = uint(14)

	IOControlNumberShift = uint(0)
	IOControlMagicShift  = IOControlNumberShift+IOControlNumberBits
	IOControlSizeShift   = IOControlMagicShift + IOControlMagicBits
	IOControlDirShift    = IOControlSizeShift + IOControlSizeBits

	IOControlWrite = uint(1)
	IOControlRead = uint(2)
)

type IOController interface {
	io.Closer
	Send(request []byte) (response []byte, err error)
}

func MakeIOControlCommand(
	direction uint,
	magic uint,
	number uint,
	size uint,
) uint {
	return direction << IOControlDirShift |
		magic <<IOControlMagicShift |
		number << IOControlNumberShift |
		size << IOControlSizeShift
}
