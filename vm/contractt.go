package vm

type CTypeUINT8 uint8
type CTypeUINT256 [32]uint8
type CTypeString string
type CTypeAddress [25]byte

var (
	CTypeUint8N   = "CTypeUINT8"
	CTypeUINT256N = "CTypeUINT256"
	CTypeStringN  = "CTypeString"
	CTypeAddressN = "CTypeString"
)
