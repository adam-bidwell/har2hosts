// domain_test.go
package main

import (
    "testing"
    "reflect"
)

func TestDomainItem_AddIP(t *testing.T) {
    tests := []struct {
        name     string
        domain   *DomainItem
        newIP    string
        expected []IPEntry
    }{
        {
            name: "Add first IP",
            domain: &DomainItem{
                domain:    "example.com",
                fromHosts: true,
                ipEntries: []IPEntry{},
            },
            newIP: "127.0.0.1",
            expected: []IPEntry{
                {ip: "127.0.0.1", selected: true},
            },
        },
        {
            name: "Add second IP",
            domain: &DomainItem{
                domain:    "example.com",
                fromHosts: true,
                ipEntries: []IPEntry{{ip: "127.0.0.1", selected: true}},
            },
            newIP: "0.0.0.0",
            expected: []IPEntry{
                {ip: "127.0.0.1", selected: true},
                {ip: "0.0.0.0", selected: true},
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.domain.ipEntries = append(tt.domain.ipEntries, IPEntry{ip: tt.newIP, selected: true})
            if !reflect.DeepEqual(tt.domain.ipEntries, tt.expected) {
                t.Errorf("got %v, want %v", tt.domain.ipEntries, tt.expected)
            }
        })
    }
}