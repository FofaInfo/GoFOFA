package gofofa

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateFields(t *testing.T) {
	// All 50 official API fields should pass individually
	allFields := []string{
		"ip", "port", "protocol", "country", "country_name",
		"region", "city", "longitude", "latitude", "asn",
		"org", "host", "domain", "os", "server",
		"icp", "title", "jarm", "header", "banner",
		"cert", "base_protocol", "link",
		"cert.issuer.org", "cert.issuer.cn", "cert.subject.org", "cert.subject.cn",
		"tls.ja3s", "tls.version",
		"cert.sn", "cert.not_before", "cert.not_after", "cert.domain",
		"header_hash", "banner_hash", "banner_fid",
		"cname", "lastupdatetime", "product", "product_category",
		"product.version", "icon_hash", "cert.is_valid", "cname_domain",
		"body", "cert.is_match", "cert.is_equal",
		"icon", "fid", "structinfo",
	}
	for _, isNext := range []bool{true, false} {
		var err error
		if isNext {
			err = ValidateFieldsNext(allFields)
		} else {
			err = ValidateFieldsAll(allFields)
		}
		assert.Nil(t, err)

		// Verify we are testing all fields in the whitelist
		if isNext {
			assert.Equal(t, len(ValidFieldsNext), len(allFields), "Test must cover all fields in ValidFieldsNext")
		} else {
			assert.Equal(t, len(ValidFieldsAll), len(allFields), "Test must cover all fields in ValidFieldsAll")
		}

		// Empty fields should pass
		if isNext {
			err = ValidateFieldsNext([]string{})
		} else {
			err = ValidateFieldsAll([]string{})
		}
		assert.Nil(t, err)

		// Invalid field should fail
		if isNext {
			err = ValidateFieldsNext([]string{"ip", "doamin"})
		} else {
			err = ValidateFieldsAll([]string{"ip", "doamin"})
		}
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "doamin")
		assert.Contains(t, err.Error(), "Unsupported export field detected")

		// Another typo
		if isNext {
			err = ValidateFieldsNext([]string{"ttle", "port"})
		} else {
			err = ValidateFieldsAll([]string{"ttle", "port"})
		}
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "ttle")

		// structinfo prefix should be rejected (only exact "structinfo" is valid)
		if isNext {
			err = ValidateFieldsNext([]string{"structinfo.any_field"})
		} else {
			err = ValidateFieldsAll([]string{"structinfo.any_field"})
		}
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "structinfo.any_field")

		// Fields removed from old whitelist should now fail
		if isNext {
			err = ValidateFieldsNext([]string{"app"})
		} else {
			err = ValidateFieldsAll([]string{"app"})
		}
		assert.NotNil(t, err)

		if isNext {
			err = ValidateFieldsNext([]string{"status_code"})
		} else {
			err = ValidateFieldsAll([]string{"status_code"})
		}
		assert.NotNil(t, err)

		if isNext {
			err = ValidateFieldsNext([]string{"type"})
		} else {
			err = ValidateFieldsAll([]string{"type"})
		}
		assert.NotNil(t, err)
	}
}
