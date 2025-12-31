package services

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"react-golang-starter/internal/models"
)

// ============ Session Service Helper Tests ============

func TestHashToken_Session(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{"refresh token", "refresh_abc123xyz"},
		{"empty token", ""},
		{"long token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIn0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := hashToken(tt.token)

			// Hash should be consistent
			if hash != hashToken(tt.token) {
				t.Error("hashToken() should return consistent results")
			}

			// Hash should be 64 chars (SHA-256 hex)
			if len(hash) != 64 {
				t.Errorf("hashToken() length = %d, want 64", len(hash))
			}
		})
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name          string
		xForwardedFor string
		xRealIP       string
		remoteAddr    string
		expectedIP    string
	}{
		{
			name:          "from X-Forwarded-For single IP",
			xForwardedFor: "192.168.1.1",
			remoteAddr:    "10.0.0.1:8080",
			expectedIP:    "192.168.1.1",
		},
		{
			name:          "from X-Forwarded-For multiple IPs",
			xForwardedFor: "192.168.1.1, 10.0.0.2, 172.16.0.1",
			remoteAddr:    "10.0.0.1:8080",
			expectedIP:    "192.168.1.1",
		},
		{
			name:       "from X-Real-IP",
			xRealIP:    "192.168.1.1",
			remoteAddr: "10.0.0.1:8080",
			expectedIP: "192.168.1.1",
		},
		{
			name:       "from RemoteAddr with port",
			remoteAddr: "192.168.1.1:8080",
			expectedIP: "192.168.1.1",
		},
		{
			name:       "from RemoteAddr without port",
			remoteAddr: "192.168.1.1",
			expectedIP: "192.168.1.1",
		},
		{
			name:       "IPv6 address",
			remoteAddr: "[::1]:8080",
			expectedIP: "[::1]",
		},
		{
			name:          "X-Forwarded-For takes precedence",
			xForwardedFor: "1.1.1.1",
			xRealIP:       "2.2.2.2",
			remoteAddr:    "3.3.3.3:8080",
			expectedIP:    "1.1.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			ip := getClientIP(req)
			if ip != tt.expectedIP {
				t.Errorf("getClientIP() = %q, want %q", ip, tt.expectedIP)
			}
		})
	}
}

func TestSessionService_ParseDeviceInfo(t *testing.T) {
	s := &SessionService{}

	tests := []struct {
		name           string
		userAgent      string
		wantDeviceType string
		wantBrowser    string
	}{
		{
			name:           "Chrome on Windows",
			userAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			wantDeviceType: "desktop",
			wantBrowser:    "Chrome",
		},
		{
			name:           "Safari on iPhone",
			userAgent:      "Mozilla/5.0 (iPhone; CPU iPhone OS 17_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Mobile/15E148 Safari/604.1",
			wantDeviceType: "mobile",
			wantBrowser:    "Safari",
		},
		{
			name:           "Safari on iPad",
			userAgent:      "Mozilla/5.0 (iPad; CPU OS 17_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Mobile/15E148 Safari/604.1",
			wantDeviceType: "mobile", // Note: useragent lib reports iPad as mobile
			wantBrowser:    "Safari",
		},
		{
			name:           "Firefox on Linux",
			userAgent:      "Mozilla/5.0 (X11; Linux x86_64; rv:120.0) Gecko/20100101 Firefox/120.0",
			wantDeviceType: "desktop",
			wantBrowser:    "Firefox",
		},
		{
			name:           "Android phone",
			userAgent:      "Mozilla/5.0 (Linux; Android 13; SM-G991B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
			wantDeviceType: "mobile",
			wantBrowser:    "Chrome",
		},
		{
			name:           "Empty user agent",
			userAgent:      "",
			wantDeviceType: "desktop",
			wantBrowser:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := s.ParseDeviceInfo(tt.userAgent)

			if info.DeviceType != tt.wantDeviceType {
				t.Errorf("ParseDeviceInfo().DeviceType = %q, want %q", info.DeviceType, tt.wantDeviceType)
			}

			if info.Browser != tt.wantBrowser {
				t.Errorf("ParseDeviceInfo().Browser = %q, want %q", info.Browser, tt.wantBrowser)
			}
		})
	}
}

func TestSessionService_GetLocationFromIP(t *testing.T) {
	s := &SessionService{}

	tests := []struct {
		name        string
		ip          string
		wantCountry string
	}{
		{
			name:        "localhost IPv4",
			ip:          "127.0.0.1",
			wantCountry: "Local",
		},
		{
			name:        "localhost IPv6",
			ip:          "::1",
			wantCountry: "Local",
		},
		{
			name:        "unknown IP",
			ip:          "8.8.8.8",
			wantCountry: "Unknown",
		},
		{
			name:        "private IP",
			ip:          "192.168.1.1",
			wantCountry: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			location := s.GetLocationFromIP(tt.ip)

			if location.Country != tt.wantCountry {
				t.Errorf("GetLocationFromIP(%q).Country = %q, want %q", tt.ip, location.Country, tt.wantCountry)
			}
		})
	}
}

func TestSessionService_GetLocationFromIP_Structure(t *testing.T) {
	s := &SessionService{}

	// Test localhost returns complete structure
	location := s.GetLocationFromIP("127.0.0.1")

	if location.CountryCode != "LO" {
		t.Errorf("GetLocationFromIP().CountryCode = %q, want %q", location.CountryCode, "LO")
	}
	if location.City != "Localhost" {
		t.Errorf("GetLocationFromIP().City = %q, want %q", location.City, "Localhost")
	}
	if location.Region != "Development" {
		t.Errorf("GetLocationFromIP().Region = %q, want %q", location.Region, "Development")
	}

	// Test unknown IP returns complete structure
	unknown := s.GetLocationFromIP("8.8.8.8")
	if unknown.CountryCode != "XX" {
		t.Errorf("GetLocationFromIP().CountryCode = %q, want %q", unknown.CountryCode, "XX")
	}
}

// ============ Device Info Structure Tests ============

func TestDeviceInfo_Fields(t *testing.T) {
	info := models.DeviceInfo{
		Browser:        "Chrome",
		BrowserVersion: "120.0.0.0",
		OS:             "Windows 10",
		OSVersion:      "10.0",
		DeviceType:     "desktop",
		DeviceName:     "Windows",
	}

	if info.Browser != "Chrome" {
		t.Errorf("DeviceInfo.Browser = %q, want %q", info.Browser, "Chrome")
	}
	if info.DeviceType != "desktop" {
		t.Errorf("DeviceInfo.DeviceType = %q, want %q", info.DeviceType, "desktop")
	}
}
