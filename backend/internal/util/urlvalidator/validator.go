package urlvalidator

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var nonPublicIPPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.88.99.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("64:ff9b::/96"),
	netip.MustParsePrefix("64:ff9b:1::/48"),
	netip.MustParsePrefix("100::/64"),
	netip.MustParsePrefix("2001::/32"),
	netip.MustParsePrefix("2001:2::/48"),
	netip.MustParsePrefix("2001:10::/28"),
	netip.MustParsePrefix("2001:20::/28"),
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("2002::/16"),
	netip.MustParsePrefix("fec0::/10"),
}

var proxyFakeIPPrefixes = []netip.Prefix{
	netip.MustParsePrefix("198.18.0.0/15"),
}

type lookupIPFunc func(context.Context, string, string) ([]net.IP, error)
type dialContextFunc func(context.Context, string, string) (net.Conn, error)

type ValidationOptions struct {
	AllowedHosts     []string
	RequireAllowlist bool
	AllowPrivate     bool
}

// ResolvedIPOptions controls the socket-level DNS validation policy. Fake-IP
// exceptions are host-scoped and only apply to the benchmarking range commonly
// used by transparent proxies. Private, loopback, and link-local answers remain
// blocked even for an allowed Fake-IP host.
type ResolvedIPOptions struct {
	AllowedFakeIPHosts []string
}

// ValidateHTTPURL validates an outbound HTTP/HTTPS URL.
//
// It provides a single validation entry point that supports:
// - scheme 校验（https 或可选允许 http）
// - 可选 allowlist（支持 *.example.com 通配）
// - allow_private_hosts 策略（阻断 localhost/私网字面量 IP）
//
// 注意：DNS Rebinding 防护（解析后 IP 校验）应在实际发起请求时执行，避免 TOCTOU。
func ValidateHTTPURL(raw string, allowInsecureHTTP bool, opts ValidationOptions) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", errors.New("url is required")
	}

	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("invalid url: %s", trimmed)
	}
	if parsed.User != nil {
		return "", errors.New("url credentials are not allowed")
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "https" && (!allowInsecureHTTP || scheme != "http") {
		return "", fmt.Errorf("invalid url scheme: %s", parsed.Scheme)
	}

	host := strings.ToLower(strings.TrimSpace(parsed.Hostname()))
	if host == "" {
		return "", errors.New("invalid host")
	}
	if !opts.AllowPrivate && isBlockedHost(host) {
		return "", fmt.Errorf("host is not allowed: %s", host)
	}

	if port := parsed.Port(); port != "" {
		num, err := strconv.Atoi(port)
		if err != nil || num <= 0 || num > 65535 {
			return "", fmt.Errorf("invalid port: %s", port)
		}
	}

	allowlist := normalizeAllowlist(opts.AllowedHosts)
	if opts.RequireAllowlist && len(allowlist) == 0 {
		return "", errors.New("allowlist is not configured")
	}
	if len(allowlist) > 0 && !isAllowedHost(host, allowlist) {
		return "", fmt.Errorf("host is not allowed: %s", host)
	}

	parsed.Path = strings.TrimRight(parsed.Path, "/")
	parsed.RawPath = ""
	return strings.TrimRight(parsed.String(), "/"), nil
}

func ValidateURLFormat(raw string, allowInsecureHTTP bool) (string, error) {
	// 最小格式校验：仅保证 URL 可解析且 scheme 合规，不做白名单/私网/SSRF 校验
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", errors.New("url is required")
	}

	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("invalid url: %s", trimmed)
	}
	if parsed.User != nil {
		return "", errors.New("url credentials are not allowed")
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "https" && (!allowInsecureHTTP || scheme != "http") {
		return "", fmt.Errorf("invalid url scheme: %s", parsed.Scheme)
	}

	host := strings.TrimSpace(parsed.Hostname())
	if host == "" {
		return "", errors.New("invalid host")
	}

	if port := parsed.Port(); port != "" {
		num, err := strconv.Atoi(port)
		if err != nil || num <= 0 || num > 65535 {
			return "", fmt.Errorf("invalid port: %s", port)
		}
	}

	return strings.TrimRight(trimmed, "/"), nil
}

func ValidateHTTPSURL(raw string, opts ValidationOptions) (string, error) {
	return ValidateHTTPURL(raw, false, opts)
}

// ValidateResolvedIP 验证 DNS 解析后的 IP 地址是否安全
// 用于防止 DNS Rebinding 攻击：在实际 HTTP 请求时调用此函数验证解析后的 IP
func ValidateResolvedIP(host string) error {
	return ValidateResolvedIPWithOptions(host, ResolvedIPOptions{})
}

func ValidateResolvedIPWithOptions(host string, opts ResolvedIPOptions) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := resolvePublicIPsWithOptions(ctx, host, net.DefaultResolver.LookupIP, opts)
	return err
}

// NewPublicDialContext returns a dial function that resolves a hostname once,
// validates every answer, and dials the validated IP directly. The original
// hostname remains on the HTTP request, so TLS certificate and SNI checks still
// use the supplier hostname.
func NewPublicDialContext(dialer *net.Dialer) func(context.Context, string, string) (net.Conn, error) {
	return NewPublicDialContextWithOptions(dialer, ResolvedIPOptions{})
}

func NewPublicDialContextWithOptions(dialer *net.Dialer, opts ResolvedIPOptions) func(context.Context, string, string) (net.Conn, error) {
	if dialer == nil {
		dialer = &net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}
	}
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		return dialPublicContextWithOptions(ctx, network, address, net.DefaultResolver.LookupIP, dialer.DialContext, opts)
	}
}

func dialPublicContext(
	ctx context.Context,
	network string,
	address string,
	lookup lookupIPFunc,
	dial dialContextFunc,
) (net.Conn, error) {
	return dialPublicContextWithOptions(ctx, network, address, lookup, dial, ResolvedIPOptions{})
}

func dialPublicContextWithOptions(
	ctx context.Context,
	network string,
	address string,
	lookup lookupIPFunc,
	dial dialContextFunc,
	opts ResolvedIPOptions,
) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	ips, err := resolvePublicIPsWithOptions(ctx, host, lookup, opts)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for _, ip := range ips {
		conn, dialErr := dial(ctx, network, net.JoinHostPort(ip.String(), port))
		if dialErr == nil {
			return conn, nil
		}
		lastErr = dialErr
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no usable public address for host %s", host)
	}
	return nil, lastErr
}

func resolvePublicIPs(ctx context.Context, host string, lookup lookupIPFunc) ([]net.IP, error) {
	return resolvePublicIPsWithOptions(ctx, host, lookup, ResolvedIPOptions{})
}

func resolvePublicIPsWithOptions(ctx context.Context, host string, lookup lookupIPFunc, opts ResolvedIPOptions) ([]net.IP, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return nil, errors.New("host is required")
	}

	var ips []net.IP
	if literal := net.ParseIP(host); literal != nil {
		ips = []net.IP{literal}
	} else {
		resolved, err := lookup(ctx, "ip", host)
		if err != nil {
			return nil, fmt.Errorf("dns resolution failed: %w", err)
		}
		ips = resolved
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("dns resolution returned no addresses for host %s", host)
	}

	allowProxyFakeIP := isAllowedProxyFakeIPHost(host, opts.AllowedFakeIPHosts)
	for _, ip := range ips {
		if !isPublicIP(ip) && !(allowProxyFakeIP && isProxyFakeIP(ip)) {
			return nil, fmt.Errorf("resolved ip %s is not allowed", ip.String())
		}
	}
	return ips, nil
}

func isAllowedProxyFakeIPHost(host string, allowedHosts []string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" || net.ParseIP(host) != nil {
		return false
	}
	return isAllowedHost(host, normalizeAllowlist(allowedHosts))
}

func isProxyFakeIP(ip net.IP) bool {
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return false
	}
	addr = addr.Unmap()
	for _, prefix := range proxyFakeIPPrefixes {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}

func isPublicIP(ip net.IP) bool {
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return false
	}
	addr = addr.Unmap()
	if !addr.IsGlobalUnicast() || addr.IsPrivate() || addr.IsLoopback() ||
		addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() || addr.IsMulticast() || addr.IsUnspecified() {
		return false
	}
	for _, prefix := range nonPublicIPPrefixes {
		if prefix.Contains(addr) {
			return false
		}
	}
	return true
}

func normalizeAllowlist(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	normalized := make([]string, 0, len(values))
	for _, v := range values {
		entry := strings.ToLower(strings.TrimSpace(v))
		if entry == "" {
			continue
		}
		if host, _, err := net.SplitHostPort(entry); err == nil {
			entry = host
		}
		normalized = append(normalized, entry)
	}
	return normalized
}

func isAllowedHost(host string, allowlist []string) bool {
	for _, entry := range allowlist {
		if entry == "" {
			continue
		}
		if strings.HasPrefix(entry, "*.") {
			suffix := strings.TrimPrefix(entry, "*.")
			if host == suffix || strings.HasSuffix(host, "."+suffix) {
				return true
			}
			continue
		}
		if host == entry {
			return true
		}
	}
	return false
}

func isBlockedHost(host string) bool {
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return !isPublicIP(ip)
	}
	return false
}
