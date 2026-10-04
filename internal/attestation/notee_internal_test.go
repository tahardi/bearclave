package attestation

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoTEESignedDigest(t *testing.T) {
	newReport := func() *Report {
		return &Report{
			Userdata:    []byte("hello world"),
			Nonce:       []byte("nonce"),
			Signature:   &Signature{R: big.NewInt(1), S: big.NewInt(2)},
			VerifyKey:   &PublicKey{X: big.NewInt(3), Y: big.NewInt(4)},
			Timestamp:   1000,
			Measurement: NoTeeMeasurement,
		}
	}

	tests := []struct {
		name      string
		mutate    func(report *Report)
		wantEqual bool
	}{
		{
			name:      "happy path - unchanged report",
			mutate:    func(_ *Report) {},
			wantEqual: true,
		},
		{
			name:      "happy path - signature is not signed",
			mutate:    func(report *Report) { report.Signature = nil },
			wantEqual: true,
		},
		{
			name:      "happy path - different signature",
			mutate:    func(report *Report) { report.Signature.R = big.NewInt(99) },
			wantEqual: true,
		},
		{
			name:   "changed userdata",
			mutate: func(report *Report) { report.Userdata = []byte("evil") },
		},
		{
			name:   "changed nonce",
			mutate: func(report *Report) { report.Nonce = []byte("other") },
		},
		{
			name:   "changed verify key x",
			mutate: func(report *Report) { report.VerifyKey.X = big.NewInt(99) },
		},
		{
			name:   "changed verify key y",
			mutate: func(report *Report) { report.VerifyKey.Y = big.NewInt(99) },
		},
		{
			name:   "changed timestamp",
			mutate: func(report *Report) { report.Timestamp++ },
		},
		{
			name:   "changed measurement",
			mutate: func(report *Report) { report.Measurement = "evil" },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			original := newReport()
			want, err := noTEESignedDigest(original)
			require.NoError(t, err)

			changed := newReport()
			tt.mutate(changed)

			// when
			got, err := noTEESignedDigest(changed)

			// then
			require.NoError(t, err)
			assert.Len(t, got, 32)
			if tt.wantEqual {
				assert.Equal(t, want, got)
				return
			}
			assert.NotEqual(t, want, got)
		})
	}

	t.Run("happy path - does not modify report", func(t *testing.T) {
		// given
		report := newReport()
		want := newReport()

		// when
		_, err := noTEESignedDigest(report)

		// then
		require.NoError(t, err)
		assert.Equal(t, want, report)
	})
}
