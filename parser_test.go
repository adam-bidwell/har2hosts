// parser_test.go
package main

import (
    "testing"
    "strings"
)

func TestParseHostsFile(t *testing.T) {
    // Create a temporary file
    hostsContent := `
# Comment line
127.0.0.1 localhost
127.0.0.1 localhost.localdomain
192.168.1.1 myserver
192.168.1.2 myserver
0.0.0.0 ads.example.com
    `

    // Mock the file reading by creating a test helper that returns this content
    domains, err := parseHostsFileContent(strings.NewReader(hostsContent))
    if err != nil {
        t.Fatalf("parseHostsFile failed: %v", err)
    }

    // Test cases
    tests := []struct {
        name           string
        domain         string
        expectedIPs    []string
        shouldExist    bool
        expectedCount  int
    }{
        {
            name:           "localhost should have one IP",
            domain:         "localhost",
            expectedIPs:    []string{"127.0.0.1"},
            shouldExist:    true,
            expectedCount:  1,
        },
        {
            name:           "myserver should have two IPs",
            domain:         "myserver",
            expectedIPs:    []string{"192.168.1.1", "192.168.1.2"},
            shouldExist:    true,
            expectedCount:  2,
        },
        {
            name:           "nonexistent domain",
            domain:         "nonexistent.com",
            expectedIPs:    nil,
            shouldExist:    false,
            expectedCount:  0,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            domain, exists := domains[tt.domain]
            if exists != tt.shouldExist {
                t.Errorf("domain existence = %v, want %v", exists, tt.shouldExist)
                return
            }

            if !tt.shouldExist {
                return
            }

            if len(domain.ipEntries) != tt.expectedCount {
                t.Errorf("got %d IPs, want %d", len(domain.ipEntries), tt.expectedCount)
                return
            }

            for i, expectedIP := range tt.expectedIPs {
                if domain.ipEntries[i].ip != expectedIP {
                    t.Errorf("IP[%d] = %s, want %s", i, domain.ipEntries[i].ip, expectedIP)
                }
            }
        })
    }
}
