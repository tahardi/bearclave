package attestation

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/google/go-sev-guest/abi"
	"github.com/google/go-sev-guest/client"
	"github.com/google/go-sev-guest/proto/sevsnp"
	"github.com/google/go-sev-guest/verify"
)

const AMD_SEV_USERDATA_SIZE = 64

type SEVAttester struct{}

func NewSEVAttester() (*SEVAttester, error) {
	return &SEVAttester{}, nil
}

func (n *SEVAttester) Attest(userdata []byte) ([]byte, error) {
	if len(userdata) > AMD_SEV_USERDATA_SIZE {
		return nil, fmt.Errorf(
			"userdata must be less than %d bytes",
			AMD_SEV_USERDATA_SIZE,
		)
	}

	sevQP, err := client.GetQuoteProvider()
	if err != nil {
		return nil, fmt.Errorf("getting sev quote provider: %w", err)
	}

	var reportData [64]byte
	copy(reportData[:], userdata)
	attestation, err := sevQP.GetRawQuote(reportData)
	if err != nil {
		return nil, fmt.Errorf("getting sev quote: %w", err)
	}
	return attestation, nil
}

type SEVVerifier struct{}

func NewSEVVerifier() (*SEVVerifier, error) {
	return &SEVVerifier{}, nil
}

// Only annoying thing is that it always returns a 64 byte slice, even if the
// userdata is less than 64 bytes.
func (n *SEVVerifier) Verify(
	report []byte,
	options ...VerifyOption,
) ([]byte, error) {
	opts := VerifyOptions{
		debug:       false,
		measurement: "",
		timestamp:   time.Now(),
	}
	for _, opt := range options {
		opt(&opts)
	}

	pbReport, err := abi.ReportCertsToProto(report)
	if err != nil {
		return nil, fmt.Errorf("converting sev report to proto: %w", err)
	}

	snpOptions := verify.DefaultOptions()
	snpOptions.Now = opts.timestamp
	err = verify.SnpAttestation(pbReport, snpOptions)
	if err != nil {
		return nil, fmt.Errorf("verifying sev report: %w", err)
	}

	err = SEVVerifyMeasurement(opts.measurement, pbReport.Report)
	if err != nil {
		return nil, fmt.Errorf("verifying measurement: %w", err)
	}

	debug, err := SEVIsDebugEnabled(pbReport.Report)
	switch {
	case err != nil:
		return nil, fmt.Errorf("getting debug mode: %w", err)
	case opts.debug != debug:
		return nil, fmt.Errorf("debug mode mismatch: expected %t, got %t",
			opts.debug,
			debug,
		)
	}
	return pbReport.Report.GetReportData(), nil
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

	got := SEVMeasurement{
		Version:         report.GetVersion(),
		GuestSVN:        report.GetGuestSvn(),
		Policy:          report.GetPolicy(),
		FamilyID:        report.GetFamilyId(),
		ImageID:         report.GetImageId(),
		VMPL:            report.GetVmpl(),
		CurrentTCB:      report.GetCurrentTcb(),
		PlatformInfo:    report.GetPlatformInfo(),
		SignerInfo:      report.GetSignerInfo(),
		Measurement:     report.GetMeasurement(),
		HostData:        report.GetHostData(),
		IDKeyDigest:     report.GetIdKeyDigest(),
		AuthorKeyDigest: report.GetAuthorKeyDigest(),
		ReportID:        report.GetReportId(),
		ReportIDMA:      report.GetReportIdMa(),
		ReportedTCB:     report.GetReportedTcb(),
		ChipID:          report.GetChipId(),
		CommittedTCB:    report.GetCommittedTcb(),
		CurrentBuild:    report.GetCurrentBuild(),
		CurrentMinor:    report.GetCurrentMinor(),
		CurrentMajor:    report.GetCurrentMajor(),
		CommittedBuild:  report.GetCommittedBuild(),
		CommittedMinor:  report.GetCommittedMinor(),
		CommittedMajor:  report.GetCommittedMajor(),
		LaunchTCB:       report.GetLaunchTcb(),
		CPUID1EAXFMS:    report.GetCpuid1EaxFms(),
	}

	if !reflect.DeepEqual(measurement, got) {
		gotJSON, err := json.Marshal(got)
		if err != nil {
			return fmt.Errorf("marshaling measurement: %w", err)
		}
		return fmt.Errorf(
			"measurement mismatch: expected '%s', got '%s'",
			measurementJSON,
			string(gotJSON),
		)
	}
	return nil
}
