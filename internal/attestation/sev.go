package attestation

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/google/go-sev-guest/abi"
	"github.com/google/go-sev-guest/client"
	"github.com/google/go-sev-guest/proto/sevsnp"
	"github.com/google/go-sev-guest/verify"
)

const AmdSevMaxUserDataSize = 64

type SEVAttester struct{}

func NewSEVAttester() (*SEVAttester, error) {
	return &SEVAttester{}, nil
}

func (n *SEVAttester) Attest(options ...AttestOption) (*AttestResult, error) {
	opts := MakeDefaultAttestOptions()
	for _, opt := range options {
		opt(&opts)
	}

	if len(opts.UserData) > AmdSevMaxUserDataSize {
		msg := fmt.Sprintf(
			"user data must be %d bytes or less",
			AmdSevMaxUserDataSize,
		)
		return nil, attesterErrorUserData(msg, nil)
	}

	sevQP, err := client.GetQuoteProvider()
	if err != nil {
		return nil, attesterError("getting sev quote provider", err)
	}

	var reportData [64]byte
	if opts.UserData != nil {
		copy(reportData[:], opts.UserData)
	}
	quote, err := sevQP.GetRawQuote(reportData)
	if err != nil {
		return nil, attesterError("getting sev quote", err)
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
	opts := MakeDefaultVerifyOptions()
	for _, opt := range options {
		opt(&opts)
	}

	pbReport, err := abi.ReportCertsToProto(attestResult.Report)
	if err != nil {
		return nil, verifierError("converting sev report to proto", err)
	}

	snpOptions := verify.DefaultOptions()
	snpOptions.Now = opts.Timestamp
	err = verify.SnpAttestation(pbReport, snpOptions)
	if err != nil {
		return nil, verifierError("verifying sev report", err)
	}

	err = SEVVerifyMeasurement(opts.Measurement, pbReport.GetReport())
	if err != nil {
		return nil, err
	}

	debug, err := SEVIsDebugEnabled(pbReport.GetReport())
	switch {
	case err != nil:
		return nil, err
	case opts.Debug != debug:
		msg := fmt.Sprintf("mode mismatch: expected %t, got %t", opts.Debug, debug)
		return nil, verifierErrorDebugMode(msg, nil)
	}

	verifyResult := &VerifyResult{
		UserData: pbReport.GetReport().GetReportData(),
	}
	return verifyResult, nil
}

func SEVIsDebugEnabled(report *sevsnp.Report) (bool, error) {
	policy, err := abi.ParseSnpPolicy(report.GetPolicy())
	if err != nil {
		return false, verifierErrorDebugMode("parsing policy", err)
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
		return verifierErrorMeasurement("unmarshaling measurement", err)
	}

	switch {
	case measurement.Version != report.GetVersion():
		msg := fmt.Sprintf(
			"version mismatch: expected %d, got %d",
			measurement.Version, report.GetVersion(),
		)
		return verifierErrorMeasurement(msg, nil)
	case measurement.GuestSVN != report.GetGuestSvn():
		msg := fmt.Sprintf(
			"guest svn mismatch: expected %d, got %d",
			measurement.GuestSVN,
			report.GetGuestSvn(),
		)
		return verifierErrorMeasurement(msg, nil)
	case measurement.Policy != report.GetPolicy():
		msg := fmt.Sprintf(
			"policy mismatch: expected %d, got %d",
			measurement.Policy,
			report.GetPolicy(),
		)
		return verifierErrorMeasurement(msg, nil)
	case !bytes.Equal(measurement.FamilyID, report.GetFamilyId()):
		msg := fmt.Sprintf(
			"family id mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.FamilyID),
			base64.StdEncoding.EncodeToString(report.GetFamilyId()),
		)
		return verifierErrorMeasurement(msg, nil)
	case !bytes.Equal(measurement.ImageID, report.GetImageId()):
		msg := fmt.Sprintf(
			"image id mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.ImageID),
			base64.StdEncoding.EncodeToString(report.GetImageId()),
		)
		return verifierErrorMeasurement(msg, nil)
	case measurement.VMPL != report.GetVmpl():
		msg := fmt.Sprintf(
			"vmpl mismatch: expected %d, got %d",
			measurement.VMPL,
			report.GetVmpl(),
		)
		return verifierErrorMeasurement(msg, nil)
	case measurement.CurrentTCB != report.GetCurrentTcb():
		msg := fmt.Sprintf(
			"current tcb mismatch: expected %d, got %d",
			measurement.CurrentTCB,
			report.GetCurrentTcb(),
		)
		return verifierErrorMeasurement(msg, nil)
	case measurement.PlatformInfo != report.GetPlatformInfo():
		msg := fmt.Sprintf(
			"platform info mismatch: expected %d, got %d",
			measurement.PlatformInfo,
			report.GetPlatformInfo(),
		)
		return verifierErrorMeasurement(msg, nil)
	case measurement.SignerInfo != report.GetSignerInfo():
		msg := fmt.Sprintf(
			"signer info mismatch: expected %d, got %d",
			measurement.SignerInfo,
			report.GetSignerInfo(),
		)
		return verifierErrorMeasurement(msg, nil)
	case !bytes.Equal(measurement.Measurement, report.GetMeasurement()):
		msg := fmt.Sprintf(
			"measurement mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.Measurement),
			base64.StdEncoding.EncodeToString(report.GetMeasurement()),
		)
		return verifierErrorMeasurement(msg, nil)
	case !bytes.Equal(measurement.HostData, report.GetHostData()):
		msg := fmt.Sprintf(
			"host data mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.HostData),
			base64.StdEncoding.EncodeToString(report.GetHostData()),
		)
		return verifierErrorMeasurement(msg, nil)
	case !bytes.Equal(measurement.IDKeyDigest, report.GetIdKeyDigest()):
		msg := fmt.Sprintf(
			"id key digest mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.IDKeyDigest),
			base64.StdEncoding.EncodeToString(report.GetIdKeyDigest()),
		)
		return verifierErrorMeasurement(msg, nil)
	case !bytes.Equal(measurement.AuthorKeyDigest, report.GetAuthorKeyDigest()):
		msg := fmt.Sprintf(
			"author key digest mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.AuthorKeyDigest),
			base64.StdEncoding.EncodeToString(report.GetAuthorKeyDigest()),
		)
		return verifierErrorMeasurement(msg, nil)
	case !bytes.Equal(measurement.ReportID, report.GetReportId()):
		msg := fmt.Sprintf(
			"report id mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.ReportID),
			base64.StdEncoding.EncodeToString(report.GetReportId()),
		)
		return verifierErrorMeasurement(msg, nil)
	case !bytes.Equal(measurement.ReportIDMA, report.GetReportIdMa()):
		msg := fmt.Sprintf(
			"report id ma mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.ReportIDMA),
			base64.StdEncoding.EncodeToString(report.GetReportIdMa()),
		)
		return verifierErrorMeasurement(msg, nil)
	case measurement.ReportedTCB != report.GetReportedTcb():
		msg := fmt.Sprintf(
			"reported tcb mismatch: expected %d, got %d",
			measurement.ReportedTCB,
			report.GetReportedTcb(),
		)
		return verifierErrorMeasurement(msg, nil)
	case !bytes.Equal(measurement.ChipID, report.GetChipId()):
		msg := fmt.Sprintf(
			"chip id mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.ChipID),
			base64.StdEncoding.EncodeToString(report.GetChipId()),
		)
		return verifierErrorMeasurement(msg, nil)
	case measurement.CommittedTCB != report.GetCommittedTcb():
		msg := fmt.Sprintf(
			"committed tcb mismatch: expected %d, got %d",
			measurement.CommittedTCB,
			report.GetCommittedTcb(),
		)
		return verifierErrorMeasurement(msg, nil)
	case measurement.CurrentBuild != report.GetCurrentBuild():
		msg := fmt.Sprintf(
			"current build mismatch: expected %d, got %d",
			measurement.CurrentBuild,
			report.GetCurrentBuild(),
		)
		return verifierErrorMeasurement(msg, nil)
	case measurement.CurrentMinor != report.GetCurrentMinor():
		msg := fmt.Sprintf(
			"current minor mismatch: expected %d, got %d",
			measurement.CurrentMinor,
			report.GetCurrentMinor(),
		)
		return verifierErrorMeasurement(msg, nil)
	case measurement.CurrentMajor != report.GetCurrentMajor():
		msg := fmt.Sprintf(
			"current major mismatch: expected %d, got %d",
			measurement.CurrentMajor,
			report.GetCurrentMajor(),
		)
		return verifierErrorMeasurement(msg, nil)
	case measurement.CommittedBuild != report.GetCommittedBuild():
		msg := fmt.Sprintf(
			"committed build mismatch: expected %d, got %d",
			measurement.CommittedBuild,
			report.GetCommittedBuild(),
		)
		return verifierErrorMeasurement(msg, nil)
	case measurement.CommittedMinor != report.GetCommittedMinor():
		msg := fmt.Sprintf(
			"committed minor mismatch: expected %d, got %d",
			measurement.CommittedMinor,
			report.GetCommittedMinor(),
		)
		return verifierErrorMeasurement(msg, nil)
	case measurement.CommittedMajor != report.GetCommittedMajor():
		msg := fmt.Sprintf(
			"committed major mismatch: expected %d, got %d",
			measurement.CommittedMajor,
			report.GetCommittedMajor(),
		)
		return verifierErrorMeasurement(msg, nil)
	case measurement.LaunchTCB != report.GetLaunchTcb():
		msg := fmt.Sprintf(
			"launch tcb mismatch: expected %d, got %d",
			measurement.LaunchTCB,
			report.GetLaunchTcb(),
		)
		return verifierErrorMeasurement(msg, nil)
	case measurement.CPUID1EAXFMS != report.GetCpuid1EaxFms():
		msg := fmt.Sprintf(
			"cpuid 1eax fms mismatch: expected %d, got %d",
			measurement.CPUID1EAXFMS,
			report.GetCpuid1EaxFms(),
		)
		return verifierErrorMeasurement(msg, nil)
	}
	return nil
}
