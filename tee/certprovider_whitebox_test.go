package tee

import (
	"crypto/x509"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMakeCertTemplate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		domain := DefaultDomain
		ip := DefaultIP
		validity := time.Hour
		now := time.Now()

		// when
		got := makeCertTemplate(domain, ip, validity, now)

		// then
		assert.True(t, got.NotBefore.Equal(now))
		assert.True(t, got.NotAfter.Equal(now.Add(validity)))
		assert.Equal(t, now.UnixNano(), got.SerialNumber.Int64())
		assert.Equal(t, domain, got.Subject.CommonName)
		assert.Equal(t, []string{domain}, got.DNSNames)
		assert.Len(t, got.IPAddresses, 1)
		assert.True(t, got.IPAddresses[0].Equal(net.ParseIP(ip)))
		assert.Equal(t, x509.KeyUsageKeyEncipherment|x509.KeyUsageDigitalSignature, got.KeyUsage)
		assert.Equal(t, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, got.ExtKeyUsage)
		assert.True(t, got.BasicConstraintsValid)
		assert.False(t, got.IsCA)
	})
}
