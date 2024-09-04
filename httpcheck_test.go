package gofofa

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckActive(t *testing.T) {
	tests := []struct {
		name          string
		fixedHostInfo string
		want          HttpResponse
	}{
		{
			name:          "Success base",
			fixedHostInfo: "http://www.baidu.com",
			want:          HttpResponse{IsActive: true, StatusCode: "200"},
		},
		{
			name:          "Fail base",
			fixedHostInfo: "http://www.sadhdkashdaskjdhsajkhkjhdaskhd.com",
			want:          HttpResponse{IsActive: false, StatusCode: "0"},
		},
		{
			name:          "IP base",
			fixedHostInfo: "123.58.224.8",
			want:          HttpResponse{IsActive: true, StatusCode: "200"},
		},
		{
			name:          "Domain base",
			fixedHostInfo: "baidu.com",
			want:          HttpResponse{IsActive: true, StatusCode: "200"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, DoHttpCheck(tt.fixedHostInfo, 3), "CheckActive(%v)", tt.fixedHostInfo)
		})
	}
}
