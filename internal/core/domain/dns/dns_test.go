package dns_test

import (
	"errors"
	"testing"
	"time"

	"github.com/sekai-labs/kumo/internal/core/domain/dns"
	"github.com/sekai-labs/kumo/internal/core/domain/zone"
)

func TestNewDNSRecord_Valid(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		zoneID    zone.ZoneID
		recType   dns.RecordType
		recName   string
		content   string
		proxied   bool
		ttl       int
		proxiable bool
	}{
		{
			name:      "valid A record proxied",
			id:        "rec-123",
			zoneID:    "zone-abc",
			recType:   dns.TypeA,
			recName:   "example.com",
			content:   "192.0.2.1",
			proxied:   true,
			ttl:       1,
			proxiable: true,
		},
		{
			name:      "valid AAAA record unproxied custom TTL",
			id:        "rec-456",
			zoneID:    "zone-abc",
			recType:   dns.TypeAAAA,
			recName:   "ipv6.example.com",
			content:   "2001:db8::1",
			proxied:   false,
			ttl:       300,
			proxiable: true,
		},
		{
			name:      "valid CNAME record proxied",
			id:        "rec-789",
			zoneID:    "zone-abc",
			recType:   dns.TypeCNAME,
			recName:   "www.example.com",
			content:   "example.com",
			proxied:   true,
			ttl:       1,
			proxiable: true,
		},
		{
			name:      "valid TXT record unproxied",
			id:        "rec-txt",
			zoneID:    "zone-abc",
			recType:   dns.TypeTXT,
			recName:   "_dmarc.example.com",
			content:   "v=DMARC1; p=reject;",
			proxied:   false,
			ttl:       3600,
			proxiable: false,
		},
		{
			name:      "valid MX record unproxied",
			id:        "rec-mx",
			zoneID:    "zone-abc",
			recType:   dns.TypeMX,
			recName:   "example.com",
			content:   "mail.example.com",
			proxied:   false,
			ttl:       1,
			proxiable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec, err := dns.NewDNSRecord(
				tt.id,
				tt.zoneID,
				tt.recType,
				tt.recName,
				tt.content,
				tt.proxied,
				tt.ttl,
				"test comment",
				nil,
				time.Now(),
			)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if rec.ID.String() != tt.id {
				t.Errorf("expected ID %q, got %q", tt.id, rec.ID.String())
			}
			if rec.Proxiable != tt.proxiable {
				t.Errorf("expected Proxiable=%v, got %v", tt.proxiable, rec.Proxiable)
			}
			if rec.Proxied != tt.proxied {
				t.Errorf("expected Proxied=%v, got %v", tt.proxied, rec.Proxied)
			}
			if int(rec.TTL) != tt.ttl {
				t.Errorf("expected TTL=%d, got %d", tt.ttl, rec.TTL)
			}
		})
	}
}

func TestNewDNSRecord_Invalid(t *testing.T) {
	validZone := zone.ZoneID("zone-123")
	now := time.Now()

	t.Run("empty record id", func(t *testing.T) {
		_, err := dns.NewDNSRecord("", validZone, dns.TypeA, "example.com", "1.1.1.1", false, 1, "", nil, now)
		if !errors.Is(err, dns.ErrEmptyRecordID) {
			t.Errorf("expected ErrEmptyRecordID, got %v", err)
		}
	})

	t.Run("empty zone id", func(t *testing.T) {
		_, err := dns.NewDNSRecord("rec-1", "", dns.TypeA, "example.com", "1.1.1.1", false, 1, "", nil, now)
		if !errors.Is(err, zone.ErrEmptyZoneID) {
			t.Errorf("expected ErrEmptyZoneID, got %v", err)
		}
	})

	t.Run("invalid record type", func(t *testing.T) {
		_, err := dns.NewDNSRecord("rec-1", validZone, "INVALID", "example.com", "1.1.1.1", false, 1, "", nil, now)
		if !errors.Is(err, dns.ErrInvalidType) {
			t.Errorf("expected ErrInvalidType, got %v", err)
		}
	})

	t.Run("empty record name", func(t *testing.T) {
		_, err := dns.NewDNSRecord("rec-1", validZone, dns.TypeA, "  ", "1.1.1.1", false, 1, "", nil, now)
		if !errors.Is(err, dns.ErrEmptyRecordName) {
			t.Errorf("expected ErrEmptyRecordName, got %v", err)
		}
	})

	t.Run("invalid A record IPv4 content", func(t *testing.T) {
		_, err := dns.NewDNSRecord("rec-1", validZone, dns.TypeA, "example.com", "999.999.999.999", false, 1, "", nil, now)
		if !errors.Is(err, dns.ErrInvalidContent) {
			t.Errorf("expected ErrInvalidContent, got %v", err)
		}
	})

	t.Run("invalid AAAA record IPv6 content", func(t *testing.T) {
		_, err := dns.NewDNSRecord("rec-1", validZone, dns.TypeAAAA, "example.com", "192.0.2.1", false, 1, "", nil, now)
		if !errors.Is(err, dns.ErrInvalidContent) {
			t.Errorf("expected ErrInvalidContent, got %v", err)
		}
	})

	t.Run("invalid CNAME record spaces", func(t *testing.T) {
		_, err := dns.NewDNSRecord("rec-1", validZone, dns.TypeCNAME, "example.com", "foo bar.com", false, 1, "", nil, now)
		if !errors.Is(err, dns.ErrInvalidContent) {
			t.Errorf("expected ErrInvalidContent, got %v", err)
		}
	})

	t.Run("empty content", func(t *testing.T) {
		_, err := dns.NewDNSRecord("rec-1", validZone, dns.TypeTXT, "example.com", "", false, 1, "", nil, now)
		if !errors.Is(err, dns.ErrInvalidContent) {
			t.Errorf("expected ErrInvalidContent, got %v", err)
		}
	})

	t.Run("invalid TTL below 60 and not 1", func(t *testing.T) {
		_, err := dns.NewDNSRecord("rec-1", validZone, dns.TypeA, "example.com", "1.1.1.1", false, 30, "", nil, now)
		if !errors.Is(err, dns.ErrInvalidTTL) {
			t.Errorf("expected ErrInvalidTTL, got %v", err)
		}
	})

	t.Run("invalid TTL above 86400", func(t *testing.T) {
		_, err := dns.NewDNSRecord("rec-1", validZone, dns.TypeA, "example.com", "1.1.1.1", false, 90000, "", nil, now)
		if !errors.Is(err, dns.ErrInvalidTTL) {
			t.Errorf("expected ErrInvalidTTL, got %v", err)
		}
	})

	t.Run("proxying non-proxiable type TXT", func(t *testing.T) {
		_, err := dns.NewDNSRecord("rec-1", validZone, dns.TypeTXT, "example.com", "test", true, 1, "", nil, now)
		if !errors.Is(err, dns.ErrCannotProxy) {
			t.Errorf("expected ErrCannotProxy, got %v", err)
		}
	})

	t.Run("proxying non-proxiable type MX", func(t *testing.T) {
		_, err := dns.NewDNSRecord("rec-1", validZone, dns.TypeMX, "example.com", "mail.example.com", true, 1, "", nil, now)
		if !errors.Is(err, dns.ErrCannotProxy) {
			t.Errorf("expected ErrCannotProxy, got %v", err)
		}
	})
}
