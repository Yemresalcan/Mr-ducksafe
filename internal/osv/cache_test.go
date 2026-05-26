package osv

import (
	"context"
	"testing"
	"time"

	"github.com/yemresalcan/mr-ducksafe/internal/matcher"
)

type stubClient struct {
	calls int
	vulns []Vulnerability
	err   error
}

func (s *stubClient) Query(_ context.Context, _ matcher.Identifier, _ string) ([]Vulnerability, error) {
	s.calls++
	return s.vulns, s.err
}

func TestCachedClient_HitDoesNotCallInner(t *testing.T) {
	stub := &stubClient{vulns: []Vulnerability{{ID: "CVE-1"}}}
	c := NewCachedClient(stub)

	id := matcher.Identifier{Type: matcher.IdentifierTypeCPE, Value: "cpe:test:test"}

	for i := 0; i < 3; i++ {
		if _, err := c.Query(context.Background(), id, "1.0"); err != nil {
			t.Fatalf("Query: %v", err)
		}
	}
	if stub.calls != 1 {
		t.Errorf("inner 1 kez çağrılmalıydı, %d kez çağrıldı", stub.calls)
	}
}

func TestCachedClient_ExpiredRecallsInner(t *testing.T) {
	stub := &stubClient{vulns: []Vulnerability{{ID: "CVE-1"}}}
	c := NewCachedClient(stub)

	clock := time.Now()
	c.now = func() time.Time { return clock }

	id := matcher.Identifier{Type: matcher.IdentifierTypeCPE, Value: "cpe:test:test"}
	if _, err := c.Query(context.Background(), id, "1.0"); err != nil {
		t.Fatal(err)
	}
	clock = clock.Add(25 * time.Hour) // TTL'i aş
	if _, err := c.Query(context.Background(), id, "1.0"); err != nil {
		t.Fatal(err)
	}
	if stub.calls != 2 {
		t.Errorf("expired sonrası 2 çağrı bekleniyor, alındı: %d", stub.calls)
	}
}
