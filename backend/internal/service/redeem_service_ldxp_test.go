package service

import "testing"

func TestIsLDXPRechargeRedeemCode(t *testing.T) {
	tests := []struct {
		name string
		code *RedeemCode
		want bool
	}{
		{
			name: "ldxp balance code",
			code: &RedeemCode{Type: RedeemTypeBalance, Value: 10, Notes: "ldxp order"},
			want: true,
		},
		{
			name: "liandong balance code",
			code: &RedeemCode{Type: RedeemTypeBalance, Value: 10, Notes: "liandong shop"},
			want: true,
		},
		{
			name: "chinese ldxp balance code",
			code: &RedeemCode{Type: RedeemTypeBalance, Value: 10, Notes: "链动小铺"},
			want: true,
		},
		{
			name: "plain balance code",
			code: &RedeemCode{Type: RedeemTypeBalance, Value: 10, Notes: "manual admin recharge"},
			want: false,
		},
		{
			name: "negative balance code",
			code: &RedeemCode{Type: RedeemTypeBalance, Value: -10, Notes: "ldxp refund"},
			want: false,
		},
		{
			name: "concurrency code",
			code: &RedeemCode{Type: RedeemTypeConcurrency, Value: 10, Notes: "ldxp order"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isLDXPRechargeRedeemCode(tt.code); got != tt.want {
				t.Fatalf("isLDXPRechargeRedeemCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMaskRedeemCode(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{name: "empty", code: "", want: ""},
		{name: "short", code: "ABC123", want: "******"},
		{name: "medium", code: "ABCDEFGH1234", want: "ABCD...34"},
		{name: "long", code: "ABCDEFGH12345678", want: "ABCDEFGH...5678"},
		{name: "trim spaces", code: "  ABCDEFGH12345678  ", want: "ABCDEFGH...5678"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := maskRedeemCode(tt.code); got != tt.want {
				t.Fatalf("maskRedeemCode() = %q, want %q", got, tt.want)
			}
		})
	}
}
