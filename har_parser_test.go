// har_parser_test.go
package main

import (
    "testing"
    "encoding/json"
    "net/url"
)

func TestParseHARDomains(t *testing.T) {
    harJSON := `{
        "log": {
            "entries": [
                {
                    "request": {
                        "url": "https://example.com/path"
                    }
                },
                {
                    "request": {
                        "url": "https://api.example.com/v1"
                    }
                },
                {
                    "request": {
                        "url": "https://example.com/other"
                    }
                }
            ]
        }
    }`

    var har HAR
    err := json.Unmarshal([]byte(harJSON), &har)
    if err != nil {
        t.Fatalf("Failed to parse HAR JSON: %v", err)
    }

    // Create domains map
    domainsMap := make(map[string]*DomainItem)
    
    // Process HAR entries
    for _, entry := range har.Log.Entries {
        if parsedURL, err := url.Parse(entry.Request.URL); err == nil {
            hostname := parsedURL.Hostname()
            if hostname != "" {
                if _, exists := domainsMap[hostname]; !exists {
                    domainsMap[hostname] = &DomainItem{
                        domain: hostname,
                        fromHosts: false,
                        ipEntries: []IPEntry{{ip: "0.0.0.0", selected: true}},
                        newEntry: true,
                    }
                }
            }
        }
    }

    // Test cases
    tests := []struct {
        name           string
        domain         string
        shouldExist    bool
        shouldBeNew    bool
        expectedIP     string
    }{
        {
            name:        "example.com should exist",
            domain:      "example.com",
            shouldExist: true,
            shouldBeNew: true,
            expectedIP:  "0.0.0.0",
        },
        {
            name:        "api.example.com should exist",
            domain:      "api.example.com",
            shouldExist: true,
            shouldBeNew: true,
            expectedIP:  "0.0.0.0",
        },
        {
            name:        "nonexistent.com should not exist",
            domain:      "nonexistent.com",
            shouldExist: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            domain, exists := domainsMap[tt.domain]
            if exists != tt.shouldExist {
                t.Errorf("domain existence = %v, want %v", exists, tt.shouldExist)
                return
            }

            if !tt.shouldExist {
                return
            }

            if domain.newEntry != tt.shouldBeNew {
                t.Errorf("newEntry = %v, want %v", domain.newEntry, tt.shouldBeNew)
            }

            if len(domain.ipEntries) != 1 {
                t.Errorf("got %d IPs, want 1", len(domain.ipEntries))
                return
            }

            if domain.ipEntries[0].ip != tt.expectedIP {
                t.Errorf("IP = %s, want %s", domain.ipEntries[0].ip, tt.expectedIP)
            }
        })
    }
}