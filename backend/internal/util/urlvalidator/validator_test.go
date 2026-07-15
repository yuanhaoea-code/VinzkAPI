package urlvalidator

import (
	"context"
	"errors"
	"net"
	"testing"
)

func TestValidateURLFormat(t *testing.T) {
	if _, err := ValidateURLFormat("", false); err == nil {
		t.Fatalf("expected empty url to fail")
	}
	if _, err := ValidateURLFormat("://bad", false); err == nil {
		t.Fatalf("expected invalid url to fail")
	}
	if _, err := ValidateURLFormat("http://example.com", false); err == nil {
		t.Fatalf("expected http to fail when allow_insecure_http is false")
	}
	if _, err := ValidateURLFormat("https://example.com", false); err != nil {
		t.Fatalf("expected https to pass, got %v", err)
	}
	if _, err := ValidateURLFormat("http://example.com", true); err != nil {
		t.Fatalf("expected http to pass when allow_insecure_http is true, got %v", err)
	}
	if _, err := ValidateURLFormat("https://example.com:bad", true); err == nil {
		t.Fatalf("expected invalid port to fail")
	}

	// 验证末尾斜杠被移除
	normalized, err := ValidateURLFormat("https://example.com/", false)
	if err != nil {
		t.Fatalf("expected trailing slash url to pass, got %v", err)
	}
	if normalized != "https://example.com" {
		t.Fatalf("expected trailing slash to be removed, got %s", normalized)
	}

	// 验证多个末尾斜杠被移除
	normalized, err = ValidateURLFormat("https://example.com///", false)
	if err != nil {
		t.Fatalf("expected multiple trailing slashes to pass, got %v", err)
	}
	if normalized != "https://example.com" {
		t.Fatalf("expected all trailing slashes to be removed, got %s", normalized)
	}

	// 验证带路径的 URL 末尾斜杠被移除
	normalized, err = ValidateURLFormat("https://example.com/api/v1/", false)
	if err != nil {
		t.Fatalf("expected trailing slash url with path to pass, got %v", err)
	}
	if normalized != "https://example.com/api/v1" {
		t.Fatalf("expected trailing slash to be removed from path, got %s", normalized)
	}
}

func TestIsPublicIPRejectsSpecialUseNetworks(t *testing.T) {
	tests := []string{
		"10.0.0.1",
		"100.64.0.1",
		"127.0.0.1",
		"169.254.169.254",
		"192.0.2.1",
		"198.18.0.1",
		"198.51.100.1",
		"203.0.113.1",
		"::1",
		"2001:db8::1",
		"fc00::1",
		"fe80::1",
	}
	for _, raw := range tests {
		t.Run(raw, func(t *testing.T) {
			if isPublicIP(net.ParseIP(raw)) {
				t.Fatalf("expected %s to be rejected as non-public", raw)
			}
		})
	}
	if !isPublicIP(net.ParseIP("1.1.1.1")) {
		t.Fatal("expected public address to pass")
	}
}

func TestDialPublicContextDialsValidatedAddressWithoutSecondLookup(t *testing.T) {
	lookupCalls := 0
	lookup := func(_ context.Context, network, host string) ([]net.IP, error) {
		lookupCalls++
		if network != "ip" || host != "supplier.example" {
			t.Fatalf("unexpected lookup %s %s", network, host)
		}
		return []net.IP{net.ParseIP("1.1.1.1")}, nil
	}

	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()
	dialedAddress := ""
	dial := func(_ context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" {
			t.Fatalf("unexpected network %s", network)
		}
		dialedAddress = address
		return clientConn, nil
	}

	conn, err := dialPublicContext(context.Background(), "tcp", "supplier.example:443", lookup, dial)
	if err != nil {
		t.Fatalf("expected dial to succeed: %v", err)
	}
	if conn != clientConn {
		t.Fatal("expected validated connection to be returned")
	}
	if lookupCalls != 1 {
		t.Fatalf("expected exactly one lookup, got %d", lookupCalls)
	}
	if dialedAddress != "1.1.1.1:443" {
		t.Fatalf("expected validated IP to be dialed, got %s", dialedAddress)
	}
}

func TestDialPublicContextRejectsMixedPublicAndPrivateAnswers(t *testing.T) {
	lookup := func(context.Context, string, string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("1.1.1.1"), net.ParseIP("127.0.0.1")}, nil
	}
	dialCalled := false
	dial := func(context.Context, string, string) (net.Conn, error) {
		dialCalled = true
		return nil, errors.New("unexpected dial")
	}

	_, err := dialPublicContext(context.Background(), "tcp", "supplier.example:443", lookup, dial)
	if err == nil {
		t.Fatal("expected mixed DNS answer to be rejected")
	}
	if dialCalled {
		t.Fatal("expected no dial after unsafe DNS answer")
	}
}

func TestValidateHTTPURL(t *testing.T) {
	if _, err := ValidateHTTPURL("http://example.com", false, ValidationOptions{}); err == nil {
		t.Fatalf("expected http to fail when allow_insecure_http is false")
	}
	if _, err := ValidateHTTPURL("http://example.com", true, ValidationOptions{}); err != nil {
		t.Fatalf("expected http to pass when allow_insecure_http is true, got %v", err)
	}
	if _, err := ValidateHTTPURL("https://example.com", false, ValidationOptions{RequireAllowlist: true}); err == nil {
		t.Fatalf("expected require allowlist to fail when empty")
	}
	if _, err := ValidateHTTPURL("https://example.com", false, ValidationOptions{AllowedHosts: []string{"api.example.com"}}); err == nil {
		t.Fatalf("expected host not in allowlist to fail")
	}
	if _, err := ValidateHTTPURL("https://api.example.com", false, ValidationOptions{AllowedHosts: []string{"api.example.com"}}); err != nil {
		t.Fatalf("expected allowlisted host to pass, got %v", err)
	}
	if _, err := ValidateHTTPURL("https://sub.api.example.com", false, ValidationOptions{AllowedHosts: []string{"*.example.com"}}); err != nil {
		t.Fatalf("expected wildcard allowlist to pass, got %v", err)
	}
	if _, err := ValidateHTTPURL("https://localhost", false, ValidationOptions{AllowPrivate: false}); err == nil {
		t.Fatalf("expected localhost to be blocked when allow_private_hosts is false")
	}
	if _, err := ValidateHTTPURL("https://127.0.0.1/v1", false, ValidationOptions{AllowPrivate: false}); err == nil {
		t.Fatalf("expected loopback address to be blocked")
	}
	if _, err := ValidateHTTPURL("https://10.0.0.1/v1", false, ValidationOptions{AllowPrivate: false}); err == nil {
		t.Fatalf("expected private address to be blocked")
	}
	if _, err := ValidateHTTPURL("https://user:pass@example.com/v1", false, ValidationOptions{}); err == nil {
		t.Fatalf("expected url credentials to be blocked")
	}
	if _, err := ValidateHTTPURL("https://new-provider.example/v1", false, ValidationOptions{AllowPrivate: false}); err != nil {
		t.Fatalf("expected arbitrary public https provider to pass without a vendor allowlist: %v", err)
	}
}
