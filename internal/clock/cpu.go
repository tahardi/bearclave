package clock

import (
	"bufio"
	"encoding/binary"
	"os"
	"strconv"
	"strings"
)

const (
	Million = 1_000_000
	Billion = 1_000_000_000

	AMD   = "AuthenticAMD"
	Intel = "GenuineIntel"

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
// read MSRs without root privileges will cause a panic. So, we attempt to
// parse the frequency from /proc/cpuinfo instead. It is both brittle and
// insecure, but we have no other solutions at the moment.
func GetTSCFrequencyAMD() (int64, error) {
	if !CheckTSCInvariant() {
		return 0, cpuErrorTSCNotInvariant("", nil)
	}

	return parseProcCPUInfo()
}

func parseProcCPUInfo() (int64, error) {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return 0, cpuErrorTSCFrequency("procinfo open", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "cpu MHz") {
			parts := strings.Split(line, ":")
			if len(parts) != 2 {
				continue
			}

			freqStr := strings.TrimSpace(parts[1])
			freqMHz, err := strconv.ParseFloat(freqStr, 64)
			if err != nil {
				continue
			}

			freqHz := int64(freqMHz * float64(Million))
			if freqHz > 0 {
				return freqHz, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, cpuErrorTSCFrequency("procinfo scan", err)
	}
	return 0, cpuErrorTSCFrequency("procinfo parse", nil)
}
