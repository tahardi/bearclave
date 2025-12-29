package clock

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	Million = 1_000_000
	Billion = 1_000_000_000

	AMD               = "AuthenticAMD"
	Intel             = "GenuineIntel"
	SystemClockSource = "/sys/devices/system/clocksource/clocksource0/current_clocksource"
	TSC               = "tsc"

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
)

//go:nosplit
func CPUID(eax, ecx uint32) (reax, rebx, recx, redx uint32)

//go:nosplit
func RDTSC() int64

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
	if baseFreqMHz != 0 {
		return int64(baseFreqMHz) * Million, nil
	}
	return CalcTSCFrequencyFromTimer()
}

// GetTSCFrequencyAMD returns the Time Stamp Counter (TSC) frequency in Hz.
// If the TSC is not invariant, an error is returned because that means it is
// not guaranteed to increase at a constant rate. Unfortunately, AMD does not
// provide TSC or processor frequency info via CPUID. Instead, they track that
// information in MSRs, which require root privileges to read. Attempting to
// read MSRs without root privileges will cause a panic. So, we calculate the
// TSC frequency by using the system time as a reference.
func GetTSCFrequencyAMD() (int64, error) {
	if !CheckTSCInvariant() {
		return 0, cpuErrorTSCNotInvariant("", nil)
	}
	return CalcTSCFrequencyFromTimer()
}

// CalcTSCFrequencyFromTimer calculates the TSC frequency by measuring the number
// of TSC ticks over a known time period. Note that we actually check and
// enforce that the system clock is the TSC. Meaning, we are essentially
// measuring the TSC with itself. While it's great that the system is using
// the secure TSC, it still uses insecure NTP to periodically sync the system
// clock (i.e., the real wall clock time). Since we assume NTP servers to be
// outside our trust boundary, we don't want to rely on the system clock.
// Instead, we will use it briefly to figure out TSC frequency and then
// implement our own clock via Network Time Security (NTS).
func CalcTSCFrequencyFromTimer() (int64, error) {
	source, err := os.ReadFile(SystemClockSource)
	switch {
	case err != nil:
		return 0, cpuErrorTSCFrequency("reading system clock source", err)
	case !strings.Contains(TSC, string(source)):
		msg := fmt.Sprintf("expected clock source '%q' got '%q'", TSC, source)
		return 0, cpuErrorTSCFrequency(msg, nil)
	}

	startTime := time.Now()
	startTSC := RDTSC()
	time.Sleep(CalibrationTime)
	endTime := time.Now()
	endTSC := RDTSC()

	elapsedNs := endTime.Sub(startTime).Nanoseconds()
	tscDelta := endTSC - startTSC

	// Calculate frequency: ticks per nanosecond * nanoseconds per second
	// TSC frequency (Hz) = (tsc_ticks / elapsed_ns) * 1_000_000_000
	if elapsedNs == 0 {
		return 0, cpuErrorTSCFrequency("calibration period too short", nil)
	}
	return (tscDelta * Billion) / elapsedNs, nil
}
