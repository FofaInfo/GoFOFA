package gofofa

import (
	"fmt"
	"strings"
)

// ValidFieldsAll is the whitelist of fields for the /search/all API endpoint.
// Based on https://fofa.info/api
var ValidFieldsAll = map[string]bool{
	"ip":               true, // IP
	"port":             true, // Port
	"protocol":         true, // Protocol
	"country":          true, // Country code
	"country_name":     true, // Country Name
	"region":           true, // Region
	"city":             true, // City
	"longitude":        true, // Longitude of geographical location
	"latitude":         true, // Latitude of geographical location
	"asn":              true, // ASN Number
	"org":              true, // ASN Organization
	"host":             true, // Host
	"domain":           true, // Domain
	"os":               true, // OS
	"server":           true, // Server
	"icp":              true, // ICP Number Information
	"title":            true, // Website Title
	"jarm":             true, // JARM Fingerprint
	"header":           true, // Type is subdomain is header
	"banner":           true, // Type is service is banner
	"cert":             true, // Cert
	"base_protocol":    true, // Base protocol, e.g. tcp/udp
	"link":             true, // Asset URL
	"cert.issuer.org":  true, // SSL issuer organization
	"cert.issuer.cn":   true, // SSL issuer common name
	"cert.subject.org": true, // SSL subject organization
	"cert.subject.cn":  true, // SSL subject common name
	"tls.ja3s":         true, // ja3s fingerprint
	"tls.version":      true, // TLS version
	"cert.sn":          true, // Certificate's serial number
	"cert.not_before":  true, // Certificate's validity start date
	"cert.not_after":   true, // Certificate's expired date
	"cert.domain":      true, // The domain name list in certificate
	"header_hash":      true, // http/https response hash value
	"banner_hash":      true, // banner response hash value
	"banner_fid":       true, // banner's structure hash value
	"cname":            true, // Domain's cname
	"lastupdatetime":   true, // FOFA last update time
	"product":          true, // Product name
	"product_category": true, // Product category
	"product.version":  true, // Product version
	"icon_hash":        true, // Favicon's hash
	"cert.is_valid":    true, // Certificate is validity or not
	"cname_domain":     true, // Domain's cname
	"body":             true, // HTML Website Body
	"cert.is_match":    true, // Certificate matches the asset's domain
	"cert.is_equal":    true, // Certificate issuer equal the certificate subject
	"icon":             true, // Icon data
	"fid":              true, // FID
	"structinfo":       true, // Structure information (partial protocol support)
}

// ValidFieldsNext is the whitelist of fields for the /search/next API endpoint.
// Based on https://fofa.info/api/batches_pages
var ValidFieldsNext = map[string]bool{
	"ip":               true, // IP
	"port":             true, // Port
	"protocol":         true, // Protocol
	"country":          true, // Country code
	"country_name":     true, // Country Name
	"region":           true, // Region
	"city":             true, // City
	"longitude":        true, // Longitude of geographical location
	"latitude":         true, // Latitude of geographical location
	"asn":              true, // ASN Number
	"org":              true, // ASN Organization
	"host":             true, // Host
	"domain":           true, // Domain
	"os":               true, // OS
	"server":           true, // Server
	"icp":              true, // ICP Number Information
	"title":            true, // Website Title
	"jarm":             true, // JARM Fingerprint
	"header":           true, // Type is subdomain is header
	"banner":           true, // Type is service is banner
	"cert":             true, // Cert
	"base_protocol":    true, // Base protocol, e.g. tcp/udp
	"link":             true, // Asset URL
	"cert.issuer.org":  true, // SSL issuer organization
	"cert.issuer.cn":   true, // SSL issuer common name
	"cert.subject.org": true, // SSL subject organization
	"cert.subject.cn":  true, // SSL subject common name
	"tls.ja3s":         true, // ja3s fingerprint
	"tls.version":      true, // TLS version
	"cert.sn":          true, // Certificate's serial number
	"cert.not_before":  true, // Certificate's validity start date
	"cert.not_after":   true, // Certificate's expired date
	"cert.domain":      true, // The domain name list in certificate
	"header_hash":      true, // http/https response hash value
	"banner_hash":      true, // banner response hash value
	"banner_fid":       true, // banner's structure hash value
	"cname":            true, // Domain's cname
	"lastupdatetime":   true, // FOFA last update time
	"product":          true, // Product name
	"product_category": true, // Product category
	"product.version":  true, // Product version
	"icon_hash":        true, // Favicon's hash
	"cert.is_valid":    true, // Certificate is validity or not
	"cname_domain":     true, // Domain's cname
	"body":             true, // HTML Website Body
	"cert.is_match":    true, // Certificate matches the asset's domain
	"cert.is_equal":    true, // Certificate issuer equal the certificate subject
	"icon":             true, // Icon data
	"fid":              true, // FID
	"structinfo":       true, // Structure information (partial protocol support)
}

// ValidateFieldsAll checks that all fields are in the ValidFieldsAll whitelist.
// Returns an error immediately if an unsupported field is detected.
func ValidateFieldsAll(fields []string) error {
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}

		if !ValidFieldsAll[f] {
			return fmt.Errorf("[Error] Unsupported export field detected: %q. Execution aborted to prevent empty results", f)
		}
	}
	return nil
}

// ValidateFieldsNext checks that all fields are in the ValidFieldsNext whitelist.
// Returns an error immediately if an unsupported field is detected.
func ValidateFieldsNext(fields []string) error {
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}

		if !ValidFieldsNext[f] {
			return fmt.Errorf("[Error] Unsupported export field detected: %q. Execution aborted to prevent empty results", f)
		}
	}
	return nil
}
