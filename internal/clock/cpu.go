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

	AMD_PSTATE_0_MSR = 0xC0010064
	FID_MASK         = 0xFF
	DID_MASK         = 0xFF00
	DID_SHIFT        = 8
	AMD_BASE_CLOCK   = 100 // MHz
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

// GetTSCFrequencyAMD Below are notes from AMD Family 17h Processors Models
// 00h-2Fh reference manual.
//
// MSRC001_0015 bit 24 TSCFreqSel. 1=TSC increments at P0 freq
// Bit 21 - LockTSCToCurrentP0, 0=TSC increments at P0, 1=TSC increments at
// P0 and will never change from this point forward even if P0 does
// Core::X86::Msr::MPERF and Msr::APERF
// 1. Write 0 to Msr::MPERF and Msr::APERF
// 2. wait for a bit
// 3. Read Msr::MPERF and Msr::APERF
// 4. P0 frequency = Msr::APERF / Msr::MPERF
// MSR0000_0010 TSC 0:63 bits contain the TSC value
// MSR0000_00E7 Max Performance Frequency Clock Count (MPERF)
// MSR0000_00E8 Actual Performance Frequency Clock Count (APERF)
func GetTSCFrequencyAMD() (int64, error) {
	if !CheckTSCInvariant() {
		return 0, cpuErrorTSCNotInvariant("", nil)
	}

	denominator, numerator, crystalHz, _ := CPUID(TscLeafEax, TscLeafEcx)
	if denominator != 0 && numerator != 0 && crystalHz != 0 {
		return int64(crystalHz) * int64(numerator) / int64(denominator), nil
	}

	return parseProcCPUInfo()
}

// parseProcCPUInfo reads /proc/cpuinfo and extracts the CPU frequency in MHz.
// It searches for the "cpu MHz" field which is present on both Intel and AMD processors.
// Returns the frequency in Hz.
func parseProcCPUInfo() (int64, error) {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return 0, cpuErrorTSCFrequency("procinfo open", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Look for "cpu MHz" field - works on both Intel and AMD
		if strings.Contains(line, "cpu MHz") {
			// Parse the format: "cpu MHz         : 4388.937"
			parts := strings.Split(line, ":")
			if len(parts) != 2 {
				continue
			}

			// Trim whitespace and parse the frequency value
			freqStr := strings.TrimSpace(parts[1])
			freqMHz, err := strconv.ParseFloat(freqStr, 64)
			if err != nil {
				continue
			}

			// Convert MHz to Hz
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
