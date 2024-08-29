package gofofa

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckActive(t *testing.T) {
	tests := []struct {
		name          string
		fixedHostInfo string
		want          bool
	}{
		// TODO: Add test cases.
		{
			name:          "Success base",
			fixedHostInfo: "http://www.baidu.com",
			want:          true,
		},
		{
			name:          "Fail base",
			fixedHostInfo: "http://www.sadhdkashdaskjdhsajkhkjhdaskhd.com",
			want:          false,
		},
		{
			name:          "IP base",
			fixedHostInfo: "123.58.224.8",
			want:          true,
		},
		{
			name:          "Domain base",
			fixedHostInfo: "baidu.com",
			want:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, CheckActive(tt.fixedHostInfo), "CheckActive(%v)", tt.fixedHostInfo)
		})
	}
}
