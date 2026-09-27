package dns

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/sekai-labs/kumo/internal/core/domain/zone"
)

var (
	ErrEmptyRecordID   = errors.New("record ID cannot be empty")
	ErrEmptyRecordName = errors.New("record name cannot be empty")
	ErrInvalidType     = errors.New("invalid or unsupported DNS record type")
	ErrInvalidContent  = errors.New("invalid DNS record content")
	ErrInvalidTTL      = errors.New("TTL must be 1 (auto) or between 60 and 86400")
	ErrCannotProxy     = errors.New("record type does not support Cloudflare proxying")
)

type RecordID string

func (id RecordID) String() string {
	return string(id)
}

func (id RecordID) Validate() error {
	if strings.TrimSpace(string(id)) == "" {
		return ErrEmptyRecordID
	}
	return nil
}

type RecordType string

const (
	TypeA     RecordType = "A"
	TypeAAAA  RecordType = "AAAA"
	TypeCNAME RecordType = "CNAME"
	TypeTXT   RecordType = "TXT"
	TypeMX    RecordType = "MX"
	TypeNS    RecordType = "NS"
	TypeSRV   RecordType = "SRV"
	TypeCAA   RecordType = "CAA"
	TypePTR   RecordType = "PTR"
)

var ValidRecordTypes = map[RecordType]bool{
	TypeA:     true,
	TypeAAAA:  true,
	TypeCNAME: true,
	TypeTXT:   true,
	TypeMX:    true,
	TypeNS:    true,
	TypeSRV:   true,
	TypeCAA:   true,
	TypePTR:   true,
}

var ProxiableTypes = map[RecordType]bool{
	TypeA:     true,
	TypeAAAA:  true,
	TypeCNAME: true,
}

type TTL int

const TTLAuto TTL = 1

func ValidateTTL(ttl int) error {
	if ttl == int(TTLAuto) {
		return nil
	}
	if ttl < 60 || ttl > 86400 {
		return ErrInvalidTTL
	}
	return nil
}

type DNSRecord struct {
	ID         RecordID    `json:"id"`
	ZoneID     zone.ZoneID `json:"zone_id"`
	Type       RecordType  `json:"type"`
	Name       string      `json:"name"`
	Content    string      `json:"content"`
	Proxied    bool        `json:"proxied"`
	TTL        TTL         `json:"ttl"`
	Proxiable  bool        `json:"proxiable"`
	Comment    string      `json:"comment,omitempty"`
	Priority   *uint16     `json:"priority,omitempty"`
	ModifiedOn time.Time   `json:"modified_on"`
}

func ValidateRecordContent(recType RecordType, content string) error {
	c := strings.TrimSpace(content)
	if c == "" {
		return ErrInvalidContent
	}
	switch recType {
	case TypeA:
		ip := net.ParseIP(c)
		if ip == nil || ip.To4() == nil {
			return fmt.Errorf("%w: '%s' is not a valid IPv4 address", ErrInvalidContent, c)
		}
	case TypeAAAA:
		ip := net.ParseIP(c)
		if ip == nil || ip.To4() != nil {
			return fmt.Errorf("%w: '%s' is not a valid IPv6 address", ErrInvalidContent, c)
		}
	case TypeCNAME:
		if strings.ContainsAny(c, " \t\r\n") {
			return fmt.Errorf("%w: CNAME content cannot contain spaces", ErrInvalidContent)
		}
	}
	return nil
}

func NewDNSRecord(
	id string,
	zoneID zone.ZoneID,
	recType RecordType,
	name string,
	content string,
	proxied bool,
	ttl int,
	comment string,
	priority *uint16,
	modifiedOn time.Time,
) (DNSRecord, error) {
	rID := RecordID(strings.TrimSpace(id))
	if err := rID.Validate(); err != nil {
		return DNSRecord{}, err
	}
	if err := zoneID.Validate(); err != nil {
		return DNSRecord{}, err
	}
	recType = RecordType(strings.ToUpper(strings.TrimSpace(string(recType))))
	if !ValidRecordTypes[recType] {
		return DNSRecord{}, ErrInvalidType
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return DNSRecord{}, ErrEmptyRecordName
	}
	if err := ValidateRecordContent(recType, content); err != nil {
		return DNSRecord{}, err
	}
	if err := ValidateTTL(ttl); err != nil {
		return DNSRecord{}, err
	}
	proxiable := ProxiableTypes[recType]
	if proxied && !proxiable {
		return DNSRecord{}, ErrCannotProxy
	}

	return DNSRecord{
		ID:         rID,
		ZoneID:     zoneID,
		Type:       recType,
		Name:       name,
		Content:    strings.TrimSpace(content),
		Proxied:    proxied,
		TTL:        TTL(ttl),
		Proxiable:  proxiable,
		Comment:    comment,
		Priority:   priority,
		ModifiedOn: modifiedOn,
	}, nil
}
