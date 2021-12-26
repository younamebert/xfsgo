package vm

type CTypeUint8 uint8
type CTypeUint16 uint16
type CTypeUint32 uint16
type CTypeUint256 [32]uint8
type CTypeString string
type CTypeAddress [25]byte

const (
	CTypeUint8N   = "CTypeUint8"
	CTypeUint16N  = "CTypeUint16"
	CTypeUint256N = "CTypeUint256"
	CTypeStringN  = "CTypeString"
	CTypeAddressN = "CTypeAddress"
)
