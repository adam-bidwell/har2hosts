// output_test.go
package main

import (
    "testing"
    "strings"
    "fmt"
)

func TestGenerateHostsFile(t *testing.T) {
    domains := []DomainItem{
        {
            domain: "localhost",
            fromHosts: true,
            ipEntries: []IPEntry{
                {ip: "127.0.0.1", selected: true},
            },
        },
        {
            domain: "myserver",
            fromHosts: true,
            ipEntries: []IPEntry{
                {ip: "192.168.1.1", selected: true},
                {ip: "192.168.1.2", selected: false},
            },
        },
        {
            domain: "ads.example.com",
            fromHosts: false,
            ipEntries: []IPEntry{
                {ip: "0.0.0.0", selected: true},
            },
            newEntry: true,
        },
    }

    output := strings.Builder{}
    output.WriteString("# Generated from HAR file: test.har\n\n")
    
    for _, item := range domains {
        for _, ipEntry := range item.ipEntries {
            if ipEntry.selected {
                output.WriteString(fmt.Sprintf("%s %s\n", ipEntry.ip, item.domain))
            }
        }
    }

    expected := `# Generated from HAR file: test.har

127.0.0.1 localhost
192.168.1.1 myserver
0.0.0.0 ads.example.com
`

    if output.String() != expected {
        t.Errorf("got:\n%s\nwant:\n%s", output.String(), expected)
    }
}