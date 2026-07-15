package service

import "testing"

func TestProductionPublicURLOrDefault(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "empty", value: "", want: DefaultPublicAPIBaseURL},
		{name: "http", value: "http://api.vinzk.cn", want: DefaultPublicAPIBaseURL},
		{name: "localhost", value: "https://localhost:3001", want: DefaultPublicAPIBaseURL},
		{name: "loopback", value: "https://127.0.0.1:3001", want: DefaultPublicAPIBaseURL},
		{name: "private", value: "https://192.168.1.8", want: DefaultPublicAPIBaseURL},
		{name: "example", value: "https://api.example.com", want: DefaultPublicAPIBaseURL},
		{name: "custom https", value: "https://gateway.vinzk.cn///", want: "https://gateway.vinzk.cn"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := productionPublicURLOrDefault(tt.value, DefaultPublicAPIBaseURL); got != tt.want {
				t.Fatalf("productionPublicURLOrDefault(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}
