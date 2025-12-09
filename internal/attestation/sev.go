package attestation

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/go-sev-guest/abi"
	"github.com/google/go-sev-guest/client"
	"github.com/google/go-sev-guest/proto/sevsnp"
	"github.com/google/go-sev-guest/verify"
)

const AmdSevMaxUserdataSize = 64

type SEVAttester struct{}

func NewSEVAttester() (*SEVAttester, error) {
	return &SEVAttester{}, nil
}

func (n *SEVAttester) Attest(options ...AttestOption) (*AttestResult, error) {
	opts := AttestOptions{
		nonce: nil,
		publicKey: nil,
		userData: nil,
	}
	for _, opt := range options {
		opt(&opts)
	}

	if len(opts.userData) > AmdSevMaxUserdataSize {
		return nil, fmt.Errorf(
			"userdata must be less than %d bytes",
			AmdSevMaxUserdataSize,
		)
	}

	sevQP, err := client.GetQuoteProvider()
	if err != nil {
		return nil, fmt.Errorf("getting sev quote provider: %w", err)
	}

	var reportData [64]byte
	if opts.userData != nil {
		copy(reportData[:], opts.userData)
	}
	quote, err := sevQP.GetRawQuote(reportData)
	if err != nil {
		return nil, fmt.Errorf("getting sev quote: %w", err)
	}
	return &AttestResult{Report: quote}, nil
}

type SEVVerifier struct{}

func NewSEVVerifier() (*SEVVerifier, error) {
	return &SEVVerifier{}, nil
}

func (n *SEVVerifier) Verify(
	attestResult *AttestResult,
	options ...VerifyOption,
) (*VerifyResult, error) {
	opts := VerifyOptions{
		debug:       false,
		measurement: "",
		timestamp:   time.Now(),
	}
	for _, opt := range options {
		opt(&opts)
	}

	pbReport, err := abi.ReportCertsToProto(attestResult.Report)
	if err != nil {
		return nil, fmt.Errorf("converting sev report to proto: %w", err)
	}

	snpOptions := verify.DefaultOptions()
	snpOptions.Now = opts.timestamp
	err = verify.SnpAttestation(pbReport, snpOptions)
	if err != nil {
		return nil, fmt.Errorf("verifying sev report: %w", err)
	}

	err = SEVVerifyMeasurement(opts.measurement, pbReport.GetReport())
	if err != nil {
		return nil, fmt.Errorf("verifying measurement: %w", err)
	}

	debug, err := SEVIsDebugEnabled(pbReport.GetReport())
	switch {
	case err != nil:
		return nil, fmt.Errorf("getting debug mode: %w", err)
	case opts.debug != debug:
		return nil, fmt.Errorf("debug mode mismatch: expected %t, got %t",
			opts.debug,
			debug,
		)
	}

	verifyResult := &VerifyResult{
		UserData: pbReport.GetReport().GetReportData(),
	}
	return verifyResult, nil
}

func SEVIsDebugEnabled(report *sevsnp.Report) (bool, error) {
	policy, err := abi.ParseSnpPolicy(report.GetPolicy())
	if err != nil {
		return false, fmt.Errorf("parsing policy: %w", err)
	}
	return policy.Debug, nil
}

type SEVMeasurement struct {
	Version         uint32 `json:"version"`
	GuestSVN        uint32 `json:"guest_svn"`
	Policy          uint64 `json:"policy"`
	FamilyID        []byte `json:"family_id"`
	ImageID         []byte `json:"image_id"`
	VMPL            uint32 `json:"vmpl"`
	CurrentTCB      uint64 `json:"current_tcb"`
	PlatformInfo    uint64 `json:"platform_info"`
	SignerInfo      uint32 `json:"signer_info"`
	Measurement     []byte `json:"measurement"`
	HostData        []byte `json:"host_data"`
	IDKeyDigest     []byte `json:"id_key_digest"`
	AuthorKeyDigest []byte `json:"author_key_digest"`
	ReportID        []byte `json:"report_id"`
	ReportIDMA      []byte `json:"report_id_ma"`
	ReportedTCB     uint64 `json:"reported_tcb"`
	ChipID          []byte `json:"chip_id"`
	CommittedTCB    uint64 `json:"committed_tcb"`
	CurrentBuild    uint32 `json:"current_build"`
	CurrentMinor    uint32 `json:"current_minor"`
	CurrentMajor    uint32 `json:"current_major"`
	CommittedBuild  uint32 `json:"committed_build"`
	CommittedMinor  uint32 `json:"committed_minor"`
	CommittedMajor  uint32 `json:"committed_major"`
	LaunchTCB       uint64 `json:"launch_tcb"`
	CPUID1EAXFMS    uint32 `json:"cpuid_1eax_fms"`
}

func SEVVerifyMeasurement(measurementJSON string, report *sevsnp.Report) error {
	if measurementJSON == "" {
		return nil
	}

	measurement := SEVMeasurement{}
	err := json.Unmarshal([]byte(measurementJSON), &measurement)
	if err != nil {
		return fmt.Errorf("unmarshaling measurement: %w", err)
	}

	switch {
	case measurement.Version != report.GetVersion():
		return fmt.Errorf("version mismatch: expected %d, got %d",
			measurement.Version,
			report.GetVersion(),
		)
	case measurement.GuestSVN != report.GetGuestSvn():
		return fmt.Errorf("guest svn mismatch: expected %d, got %d",
			measurement.GuestSVN,
			report.GetGuestSvn(),
		)
	case measurement.Policy != report.GetPolicy():
		return fmt.Errorf("policy mismatch: expected %d, got %d",
			measurement.Policy,
			report.GetPolicy(),
		)
	case !bytes.Equal(measurement.FamilyID, report.GetFamilyId()):
		return fmt.Errorf("family id mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.FamilyID),
			base64.StdEncoding.EncodeToString(report.GetFamilyId()),
		)
	case !bytes.Equal(measurement.ImageID, report.GetImageId()):
		return fmt.Errorf("image id mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.ImageID),
			base64.StdEncoding.EncodeToString(report.GetImageId()),
		)
	case measurement.VMPL != report.GetVmpl():
		return fmt.Errorf("vmpl mismatch: expected %d, got %d",
			measurement.VMPL,
			report.GetVmpl(),
		)
	case measurement.CurrentTCB != report.GetCurrentTcb():
		return fmt.Errorf("current tcb mismatch: expected %d, got %d",
			measurement.CurrentTCB,
			report.GetCurrentTcb(),
		)
	case measurement.PlatformInfo != report.GetPlatformInfo():
		return fmt.Errorf("platform info mismatch: expected %d, got %d",
			measurement.PlatformInfo,
			report.GetPlatformInfo(),
		)
	case measurement.SignerInfo != report.GetSignerInfo():
		return fmt.Errorf("signer info mismatch: expected %d, got %d",
			measurement.SignerInfo,
			report.GetSignerInfo(),
		)
	case !bytes.Equal(measurement.Measurement, report.GetMeasurement()):
		return fmt.Errorf("measurement mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.Measurement),
			base64.StdEncoding.EncodeToString(report.GetMeasurement()),
		)
	case !bytes.Equal(measurement.HostData, report.GetHostData()):
		return fmt.Errorf("host data mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.HostData),
			base64.StdEncoding.EncodeToString(report.GetHostData()),
		)
	case !bytes.Equal(measurement.IDKeyDigest, report.GetIdKeyDigest()):
		return fmt.Errorf("id key digest mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.IDKeyDigest),
			base64.StdEncoding.EncodeToString(report.GetIdKeyDigest()),
		)
	case !bytes.Equal(measurement.AuthorKeyDigest, report.GetAuthorKeyDigest()):
		return fmt.Errorf("author key digest mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.AuthorKeyDigest),
			base64.StdEncoding.EncodeToString(report.GetAuthorKeyDigest()),
		)
	case !bytes.Equal(measurement.ReportID, report.GetReportId()):
		return fmt.Errorf("report id mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.ReportID),
			base64.StdEncoding.EncodeToString(report.GetReportId()),
		)
	case !bytes.Equal(measurement.ReportIDMA, report.GetReportIdMa()):
		return fmt.Errorf("report id ma mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.ReportIDMA),
			base64.StdEncoding.EncodeToString(report.GetReportIdMa()),
		)
	case measurement.ReportedTCB != report.GetReportedTcb():
		return fmt.Errorf("reported tcb mismatch: expected %d, got %d",
			measurement.ReportedTCB,
			report.GetReportedTcb(),
		)
	case !bytes.Equal(measurement.ChipID, report.GetChipId()):
		return fmt.Errorf("chip id mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.ChipID),
			base64.StdEncoding.EncodeToString(report.GetChipId()),
		)
	case measurement.CommittedTCB != report.GetCommittedTcb():
		return fmt.Errorf("committed tcb mismatch: expected %d, got %d",
			measurement.CommittedTCB,
			report.GetCommittedTcb(),
		)
	case measurement.CurrentBuild != report.GetCurrentBuild():
		return fmt.Errorf("current build mismatch: expected %d, got %d",
			measurement.CurrentBuild,
			report.GetCurrentBuild(),
		)
	case measurement.CurrentMinor != report.GetCurrentMinor():
		return fmt.Errorf("current minor mismatch: expected %d, got %d",
			measurement.CurrentMinor,
			report.GetCurrentMinor(),
		)
	case measurement.CurrentMajor != report.GetCurrentMajor():
		return fmt.Errorf("current major mismatch: expected %d, got %d",
			measurement.CurrentMajor,
			report.GetCurrentMajor(),
		)
	case measurement.CommittedBuild != report.GetCommittedBuild():
		return fmt.Errorf("committed build mismatch: expected %d, got %d",
			measurement.CommittedBuild,
			report.GetCommittedBuild(),
		)
	case measurement.CommittedMinor != report.GetCommittedMinor():
		return fmt.Errorf("committed minor mismatch: expected %d, got %d",
			measurement.CommittedMinor,
			report.GetCommittedMinor(),
		)
	case measurement.CommittedMajor != report.GetCommittedMajor():
		return fmt.Errorf("committed major mismatch: expected %d, got %d",
			measurement.CommittedMajor,
			report.GetCommittedMajor(),
		)
	case measurement.LaunchTCB != report.GetLaunchTcb():
		return fmt.Errorf("launch tcb mismatch: expected %d, got %d",
			measurement.LaunchTCB,
			report.GetLaunchTcb(),
		)
	case measurement.CPUID1EAXFMS != report.GetCpuid1EaxFms():
		return fmt.Errorf("cpuid 1eax fms mismatch: expected %d, got %d",
			measurement.CPUID1EAXFMS,
			report.GetCpuid1EaxFms(),
		)
	}
	return nil
}
