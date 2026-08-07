package gofofa

import "testing"

func TestAPIResponseError(t *testing.T) {
	tests := []struct {
		name     string
		failed   bool
		errmsg   string
		fallback string
		want     string
	}{
		{
			name:   "success",
			errmsg: "ignored informational text",
		},
		{
			name:   "API message",
			failed: true,
			errmsg: "[820000] 查询语法错误",
			want:   "[820000] 查询语法错误",
		},
		{
			name:     "fallback",
			failed:   true,
			fallback: "fofa request failed",
			want:     "fofa request failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := apiResponseError(tt.failed, tt.errmsg, tt.fallback)
			if tt.want == "" {
				if err != nil {
					t.Fatalf("apiResponseError() = %v, want nil", err)
				}
				return
			}
			if err == nil || err.Error() != tt.want {
				t.Fatalf("apiResponseError() = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestDecodeAPIErrorEnvelope(t *testing.T) {
	response, err := decodeAPIErrorEnvelope([]byte(`{"error":true,"errmsg":"[45012] 请求速度过快"}`))
	if err != nil {
		t.Fatalf("decodeAPIErrorEnvelope() error = %v", err)
	}
	if !response.Error || response.Errmsg != "[45012] 请求速度过快" {
		t.Fatalf("decodeAPIErrorEnvelope() = %+v", response)
	}
}
