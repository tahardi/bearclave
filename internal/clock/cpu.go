package clock

import (
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

const (
	Million = 1_000_000
	Billion = 1_000_000_000

	AMD   = "AuthenticAMD"
	Intel = "GenuineIntel"

	CalibrationTime = 100 * time.Millisecond

	InvariantBit        = 1 << 8
	InvariantTscLeafEax = 0x80000007
	InvariantTscLeafEcx = 0
	ProcFreqLeafEax     = 0x16
	ProcFreqLeafEcx     = 0
	TscLeafEax          = 0x15
	TscLeafEcx          = 0
	VendorLeafEax       = 0
	VendorLeafEcx       = 0
	VendorLength        = 12

	AmdGuestTscScaleOffset  = 0x2F0
	AmdGuestTscOffsetOffset = 0x2F8
	AmdSevFeaturesOffset    = 0x3B0
	AmdSecureTscEnabledBit  = 1 << 11 //Pg.750 arch. progr. man.
	AmdGuestTscFrequencyMsr = 0xC0010134
)

//go:nosplit
func CPUID(eax, ecx uint32) (reax, rebx, recx, redx uint32)

//go:nosplit
func RDTSC() int64

//go:nosplit
func RDMSR(msr uint32) uint64

func CheckTSCInvariant() bool {
	_, _, _, edx := CPUID(InvariantTscLeafEax, InvariantTscLeafEcx)
	return edx&InvariantBit > 0
}

func GetVendor() string {
	// Vendor string is in EBX, EDX, ECX (in that order)
	_, ebx, ecx, edx := CPUID(VendorLeafEax, VendorLeafEcx)
	vendorBytes := make([]byte, VendorLength)
	binary.LittleEndian.PutUint32(vendorBytes[0:4], ebx)
	binary.LittleEndian.PutUint32(vendorBytes[4:8], edx)
	binary.LittleEndian.PutUint32(vendorBytes[8:12], ecx)
	return string(vendorBytes)
}

func GetTSCFrequency() (int64, error) {
	vendor := GetVendor()
	switch vendor {
	case AMD:
		return GetTSCFrequencyAMD()
	case Intel:
		return GetTSCFrequencyIntel()
	default:
		return 0, cpuErrorVendor(vendor, nil)
	}
}

// GetTSCFrequencyIntel returns the Time Stamp Counter (TSC) frequency in Hz.
// If the TSC is not invariant, an error is returned because that means it is
// not guaranteed to increase at a constant rate. We attempt to calculate the
// TSC frequency with information from the TSC leaf. If that information is
// incomplete, we fall back to calculating TSC frequency from the processor's
// base frequency. Specification for CPUID command and registers found at:
//
// https://www.felixcloutier.com/x86/cpuid
func GetTSCFrequencyIntel() (int64, error) {
	if !CheckTSCInvariant() {
		return 0, cpuErrorTSCNotInvariant("", nil)
	}

	denominator, numerator, crystalHz, _ := CPUID(TscLeafEax, TscLeafEcx)
	if denominator != 0 && numerator != 0 && crystalHz != 0 {
		return int64(crystalHz) * int64(numerator) / int64(denominator), nil
	}

	baseFreqMHz, _, _, _ := CPUID(ProcFreqLeafEax, ProcFreqLeafEcx)
	baseFreqMHz &= 0xFFFF
	if baseFreqMHz == 0 {
		return 0, cpuErrorTSCFrequency("Intel", nil)
	}
	return int64(baseFreqMHz) * Million, nil
}

// GetTSCFrequencyAMD returns the Time Stamp Counter (TSC) frequency in Hz.
// If the TSC is not invariant, an error is returned because that means it is
// not guaranteed to increase at a constant rate. Unfortunately, AMD does not
// provide TSC or processor frequency info via CPUID. Instead, they track that
// information in MSRs, which require root privileges to read. Attempting to
// read MSRs without root privileges will cause a panic. So, we calculate the
// TSC frequency by using the system time as a reference. This is exactly
// what we were hoping to avoid, as we consider the system time to be untrusted.
func GetTSCFrequencyAMD() (int64, error) {
	if !CheckTSCInvariant() {
		return 0, cpuErrorTSCNotInvariant("", nil)
	}

	// TODO: Consider checking SEV_FEATURES for SecureTSC enabled
	// SEV_FEATURES_OFFSET = 0x3B0h - bit 11 SecureTSC. Pg.750 arch. progr. man.

	// TODO: I don't know if reading this MSR requires root privileges. If not,
	// you could read the TSC frequency from it.
	// Guests may read GUEST_TSC_FREQ from MSR (C001_0134h) to get freq in MHz
	//scalingInfo, err := readTSCScalingFromMSR()
	//if err != nil {
	//	return 0, err
	//}
	//
	//fmt.Printf("Read scaling info: %x, %x\n", scalingInfo.Scale, scalingInfo.Offset)
	//freq, err := readTSCFrequencyFromMSR()
	//if err != nil {
	//	return 0, err
	//}
	//fmt.Printf("Read MSR freq: %d\n", freq)
	source, err := os.ReadFile("/sys/devices/system/clocksource/clocksource0/current_clocksource")
	if err == nil {
		fmt.Printf("Current clocksource: %s", source)
	}
	return CalcTSCFrequencyFromTimer(CalibrationTime)
}

func readTSCFrequencyFromMSR() (uint64, error) {
	return RDMSR(AmdGuestTscFrequencyMsr), nil
}

// TSCScalingInfo holds the TSC scaling parameters from /dev/cpu/CPUID/msr
type TSCScalingInfo struct {
	Scale  uint64 // GUEST_TSC_SCALE: 32-bit scale factor (high 32 bits of 64-bit MSR)
	Offset uint64  // GUEST_TSC_OFFSET: 64-bit offset
}

// Pg. 749 AMD64 Architecture Programmer's Manual Vol 2: System Programming (24593)
// specifies where the GUEST_TSC_SCALE and GUEST_TSC_OFFSET values are located
// in VMSA layout.
func readTSCScalingFromMSR() (*TSCScalingInfo, error) {
	file, err := os.Open("/dev/cpu/0/msr")
	if err != nil {
		return nil, cpuErrorTSCFrequency("opening /dev/cpu/0/msr", err)
	}
	defer file.Close()

	_, err = file.Seek(AmdGuestTscFrequencyMsr, 0)
	if err != nil {
		return nil, cpuErrorTSCFrequency("seeking to MSR_GUEST_TSC_FREQ", err)
	}

	freqBuffer := make([]byte, 8)
	n, err := file.Read(freqBuffer)
	if err != nil || n != 8 {
		return nil, cpuErrorTSCFrequency("reading MSR_GUEST_TSC_FREQ", err)
	}
	freq := binary.LittleEndian.Uint64(freqBuffer)
	fmt.Printf("MSR freq: %d\n", freq)

	_, err = file.Seek(AmdGuestTscScaleOffset, 0)
	if err != nil {
		return nil, cpuErrorTSCFrequency("seeking to GUEST_TSC_SCALE", err)
	}

	// GUEST_TSC_SCALE - 8.32 fixed point binary number (8-bit int, 32-bit fraction)
	scaleBuffer := make([]byte, 8)
	n, err = file.Read(scaleBuffer)
	if err != nil || n != 8 {
		return nil, cpuErrorTSCFrequency("reading GUEST_TSC_SCALE", err)
	}

	_, err = file.Seek(AmdGuestTscOffsetOffset, 0)
	if err != nil {
		return nil, cpuErrorTSCFrequency("seeking to GUEST_TSC_OFFSET", err)
	}

	offsetBuffer := make([]byte, 8)
	n, err = file.Read(offsetBuffer)
	if err != nil || n != 8 {
		return nil, cpuErrorTSCFrequency("reading GUEST_TSC_OFFSET", err)
	}

	_, err = file.Seek(AmdSevFeaturesOffset, 0)
	if err != nil {
		return nil, cpuErrorTSCFrequency("seeking to SEV_FEATURES", err)
	}

	featuresBuffer := make([]byte, 8)
	n, err = file.Read(featuresBuffer)
	if err != nil || n != 8 {
		return nil, cpuErrorTSCFrequency("reading SEV_FEATURES", err)
	}

	scale := binary.LittleEndian.Uint64(scaleBuffer)
	offset := binary.LittleEndian.Uint64(offsetBuffer)
	features := binary.LittleEndian.Uint64(featuresBuffer)
	if features&AmdSecureTscEnabledBit == 0 {
		return nil, cpuErrorTSCFrequency("SEV SecureTSC not enabled", nil)
	}
	return &TSCScalingInfo{Scale: scale, Offset: offset}, nil
}


// CalcTSCFrequencyFromTimer calculates the TSC frequency by measuring the number
// of TSC ticks over a known time period.
func CalcTSCFrequencyFromTimer(duration time.Duration) (int64, error) {
	startTime := time.Now()
	startTSC := RDTSC()

	time.Sleep(duration)

	endTime := time.Now()
	endTSC := RDTSC()

	elapsedNs := endTime.Sub(startTime).Nanoseconds()
	tscDelta := endTSC - startTSC

	// Calculate frequency: ticks per nanosecond * nanoseconds per second
	// TSC frequency (Hz) = (tsc_ticks / elapsed_ns) * 1_000_000_000
	if elapsedNs == 0 {
		return 0, cpuErrorTSCFrequency("AMD", nil)
	}
	return (tscDelta * Billion) / elapsedNs, nil
}
