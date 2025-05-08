package sdk

import (
	"fmt"
	"github.com/tahardi/bearclave"
)

func MakeAttester(platform Platform) (bearclave.Attester, error) {
	switch platform {
	case Nitro:
		return bearclave.NewNitroAttester()
	case SEV:
		return bearclave.NewSEVAttester()
	case TDX:
		return bearclave.NewTDXAttester()
	case Unsafe:
		return bearclave.NewUnsafeAttester()
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}

func MakeVerifier(platform Platform) (bearclave.Verifier, error) {
	switch platform {
	case Nitro:
		return bearclave.NewNitroVerifier()
	case SEV:
		return bearclave.NewSEVVerifier()
	case TDX:
		return bearclave.NewTDXVerifier()
	case Unsafe:
		return bearclave.NewUnsafeVerifier()
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}

func MakeTransporter(
	platform Platform,
	sendCID int,
	sendPort int,
	receivePort int,
) (bearclave.Transporter, error) {
	switch platform {
	case Nitro:
		return bearclave.NewNitroTransporter(sendCID, sendPort, receivePort)
	case SEV:
		return bearclave.NewSEVTransporter(sendPort, receivePort)
	case TDX:
		return bearclave.NewTDXTransporter(sendPort, receivePort)
	case Unsafe:
		return bearclave.NewUnsafeTransporter(sendPort, receivePort)
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}
