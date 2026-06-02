//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

const (
	spdxLicensesURL   = "https://spdx.org/licenses/licenses.json"
	spdxExceptionsURL = "https://spdx.org/licenses/exceptions.json"
)

type license struct {
	LicenseID string `json:"licenseId"`
	Name      string `json:"name"`
}

type licensesDocument struct {
	Licenses []license `json:"licenses"`
}

type exception struct {
	LicenseExceptionID string `json:"licenseExceptionId"`
	Name               string `json:"name"`
}

type exceptionsDocument struct {
	Exceptions []exception `json:"exceptions"`
}

type output struct {
	Licenses   map[string]string `json:"licenses"`
	Exceptions map[string]string `json:"exceptions"`
}

func fetchJSON(url string, into any) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("fetching %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetching %s: unexpected status %s", url, resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(into); err != nil {
		return fmt.Errorf("decoding %s: %w", url, err)
	}
	return nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <output-path>\n", os.Args[0])
		os.Exit(1)
	}
	outputPath := os.Args[1]

	var licensesDoc licensesDocument
	if err := fetchJSON(spdxLicensesURL, &licensesDoc); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var exceptionsDoc exceptionsDocument
	if err := fetchJSON(spdxExceptionsURL, &exceptionsDoc); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	out := output{
		Licenses:   make(map[string]string, len(licensesDoc.Licenses)),
		Exceptions: make(map[string]string, len(exceptionsDoc.Exceptions)),
	}
	for _, l := range licensesDoc.Licenses {
		out.Licenses[l.LicenseID] = l.Name
	}
	for _, e := range exceptionsDoc.Exceptions {
		out.Exceptions[e.LicenseExceptionID] = e.Name
	}

	f, err := os.Create(outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "creating %s: %v\n", outputPath, err)
		os.Exit(1)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "encoding output: %v\n", err)
		os.Exit(1)
	}
}
