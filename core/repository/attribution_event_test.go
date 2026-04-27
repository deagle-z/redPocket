package repository

import "testing"

func TestValidateAttributionEventName(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "trim valid", in: "  page.view  ", want: "page.view"},
		{name: "underscore hyphen dot", in: "signup_submit-v2.ok", want: "signup_submit-v2.ok"},
		{name: "empty", in: "   ", wantErr: true},
		{name: "too long", in: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", wantErr: true},
		{name: "invalid char", in: "page view", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateAttributionEventName(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ValidateAttributionEventName() expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ValidateAttributionEventName() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("ValidateAttributionEventName() = %q, want %q", got, tt.want)
			}
		})
	}
}
