package spdx

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:generate go run ../../scripts/gen-spdx.go spdx.json

//go:embed spdx.json
var spdxData []byte

type spdxDocument struct {
	Licenses   map[string]string `json:"licenses"`
	Exceptions map[string]string `json:"exceptions"`
}

var (
	licenseNames   map[string]string
	exceptionNames map[string]string
)

func init() {
	licenseNames, exceptionNames = mustDecode(spdxData)
}

// mustDecode parses an SPDX document from data and returns its license and
// exception name tables. It panics with an error wrapping the underlying
// decoding failure when data is not a valid SPDX document.
func mustDecode(data []byte) (licenses, exceptions map[string]string) {
	var doc spdxDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		panic(fmt.Errorf("spdx: parsing embedded data: %w", err))
	}
	return doc.Licenses, doc.Exceptions
}

// LookupName returns the human-readable name registered for id, which may be
// either an SPDX license identifier or an SPDX exception identifier. The
// second return value is false when id is unknown.
func LookupName(id string) (string, bool) {
	if name, ok := licenseNames[id]; ok {
		return name, true
	}
	if name, ok := exceptionNames[id]; ok {
		return name, true
	}
	return "", false
}

// IsKnownIdentifier reports whether id is a known SPDX license or exception
// identifier.
func IsKnownIdentifier(id string) bool {
	return IsKnownLicenseID(id) || IsKnownExceptionID(id)
}

// IsKnownLicenseID reports whether id is a known SPDX license identifier.
func IsKnownLicenseID(id string) bool {
	_, ok := licenseNames[id]
	return ok
}

// IsKnownExceptionID reports whether id is a known SPDX exception identifier.
func IsKnownExceptionID(id string) bool {
	_, ok := exceptionNames[id]
	return ok
}
