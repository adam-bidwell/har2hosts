package main

import (
    "encoding/json"
    "fmt"
    "net/url"
    "os"
    "sort"
    "strings"
    "bufio"
    "github.com/gdamore/tcell/v2"
)

// HAR file structures
type HAR struct {
    Log Log `json:"log"`
}

type Log struct {
    Entries []Entry `json:"entries"`
}

type Entry struct {
    Request Request `json:"request"`
}

type Request struct {
    URL string `json:"url"`
}

// IPEntry represents a single IP address mapping for a domain
type IPEntry struct {
    ip       string
    selected bool
}

// DomainItem represents a domain with all its IP addresses and sources
type DomainItem struct {
    domain    string
    fromHosts bool              // true if from hosts file
    ipEntries []IPEntry        // all IP addresses for this domain
    newEntry  bool             // true if this is a new entry from HAR
}

func parseHostsFile() (map[string]*DomainItem, error) {
    domains := make(map[string]*DomainItem)
    
    file, err := os.Open("/etc/hosts")
    if err != nil {
        return domains, err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        // Skip comments and empty lines
        if strings.HasPrefix(line, "#") || line == "" {
            continue
        }
        
        // Split on whitespace
        fields := strings.Fields(line)
        if len(fields) >= 2 {
            domain := fields[1]
            ip := fields[0]
            
            if item, exists := domains[domain]; exists {
                // Add additional IP for existing domain
                item.ipEntries = append(item.ipEntries, IPEntry{ip: ip, selected: true})
            } else {
                // Create new domain entry
                domains[domain] = &DomainItem{
                    domain: domain,
                    fromHosts: true,
                    ipEntries: []IPEntry{{ip: ip, selected: true}},
                    newEntry: false,
                }
            }
        }
    }
    
    return domains, scanner.Err()
}

func drawList(s tcell.Screen, domains []DomainItem, currentIdx int) {
    s.Clear()
    width, height := s.Size()
    defaultStyle := tcell.StyleDefault
    selectedStyle := defaultStyle.Background(tcell.ColorGray)
    hostsStyle := defaultStyle.Foreground(tcell.ColorYellow)
    hostsSelectedStyle := selectedStyle.Foreground(tcell.ColorYellow)

    // Draw title
    title := "Space to toggle selection, Enter to save, Esc to quit | White=HAR Yellow=Hosts"
    for i, r := range title {
        s.SetContent(i, 0, r, nil, defaultStyle)
    }

    // Calculate visible range
    visibleStart := 0
    if currentIdx > height-3 {
        visibleStart = currentIdx - (height - 3)
    }
    visibleEnd := min(visibleStart+height-2, len(domains))

    // Draw domains
    for i := visibleStart; i < visibleEnd; i++ {
        y := i - visibleStart + 1
        
        // Choose style based on source and selection
        var lineStyle tcell.Style
        if i == currentIdx {
            if domains[i].fromHosts {
                lineStyle = hostsSelectedStyle
            } else {
                lineStyle = selectedStyle
            }
        } else {
            if domains[i].fromHosts {
                lineStyle = hostsStyle
            } else {
                lineStyle = defaultStyle
            }
        }

        // Format the line based on number of IPs
        var line string
        if domains[i].newEntry {
            prefix := "[ ]"
            if domains[i].ipEntries[0].selected {
                prefix = "[✓]"
            }
            line = fmt.Sprintf("%s %s (new: 0.0.0.0)", prefix, domains[i].domain)
        } else if len(domains[i].ipEntries) == 1 {
            prefix := "[ ]"
            if domains[i].ipEntries[0].selected {
                prefix = "[✓]"
            }
            line = fmt.Sprintf("%s %s (%s)", prefix, domains[i].domain, domains[i].ipEntries[0].ip)
        } else {
            // Multiple IPs - show count and details
            selectedCount := 0
            for _, ip := range domains[i].ipEntries {
                if ip.selected {
                    selectedCount++
                }
            }
            line = fmt.Sprintf("[%d/%d] %s (%d IPs: ", selectedCount, len(domains[i].ipEntries), 
                              domains[i].domain, len(domains[i].ipEntries))
            
            // Add abbreviated IP list
            for j, ip := range domains[i].ipEntries {
                if j > 0 {
                    line += ", "
                }
                if ip.selected {
                    line += "*"
                }
                line += ip.ip
            }
            line += ")"
        }

        // Write the line
        for x, r := range line {
            if x < width {  // Prevent writing past screen width
                s.SetContent(x, y, r, nil, lineStyle)
            }
        }
    }

    s.Show()
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

func main() {
    if len(os.Args) != 2 {
        fmt.Println("Usage: harToHosts <harfile>")
        os.Exit(1)
    }

    harFile := os.Args[1]

    // Read and parse HAR file
    data, err := os.ReadFile(harFile)
    if err != nil {
        fmt.Printf("Error reading file: %v\n", err)
        os.Exit(1)
    }

    var har HAR
    if err := json.Unmarshal(data, &har); err != nil {
        fmt.Printf("Error parsing HAR JSON: %v\n", err)
        os.Exit(1)
    }

    // Get existing hosts domains
    domainsMap, err := parseHostsFile()
    if err != nil {
        fmt.Printf("Warning: Error reading hosts file: %v\n", err)
    }

    // Add HAR domains
    for _, entry := range har.Log.Entries {
        if parsedURL, err := url.Parse(entry.Request.URL); err == nil {
            hostname := parsedURL.Hostname()
            if hostname != "" {
                if _, exists := domainsMap[hostname]; !exists {
                    // Create new domain entry with 0.0.0.0
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

    // Convert to sorted slice
    var domains []DomainItem
    for _, item := range domainsMap {
        domains = append(domains, *item)
    }
    sort.Slice(domains, func(i, j int) bool {
        return domains[i].domain < domains[j].domain
    })

    // Initialize screen
    s, err := tcell.NewScreen()
    if err != nil {
        fmt.Printf("Error creating screen: %v\n", err)
        os.Exit(1)
    }
    if err := s.Init(); err != nil {
        fmt.Printf("Error initializing screen: %v\n", err)
        os.Exit(1)
    }
    defer s.Fini()

    // Main event loop
    currentIdx := 0
    for {
        drawList(s, domains, currentIdx)

        switch ev := s.PollEvent().(type) {
        case *tcell.EventKey:
            switch ev.Key() {
            case tcell.KeyEscape:
                return
            case tcell.KeyEnter:
                // Generate hosts file with selected domains
                output := strings.Builder{}
                output.WriteString("# Generated from HAR file: " + harFile + "\n\n")
                
                // First write all hosts file entries with their original IPs
                for _, item := range domains {
                    for _, ipEntry := range item.ipEntries {
                        if ipEntry.selected {
                            output.WriteString(fmt.Sprintf("%s %s\n", ipEntry.ip, item.domain))
                        }
                    }
                }
                
                outputFile := "hosts.txt"
                if err := os.WriteFile(outputFile, []byte(output.String()), 0644); err != nil {
                    s.Fini()
                    fmt.Printf("Error writing hosts file: %v\n", err)
                    os.Exit(1)
                }
                
                s.Fini()
                fmt.Printf("Successfully created %s\n", outputFile)
                return
            case tcell.KeyUp:
                if currentIdx > 0 {
                    currentIdx--
                }
            case tcell.KeyDown:
                if currentIdx < len(domains)-1 {
                    currentIdx++
                }
            case tcell.KeyRune:
                switch ev.Rune() {
                case ' ':
                    // Toggle all IP entries for the domain
                    for i := range domains[currentIdx].ipEntries {
                        domains[currentIdx].ipEntries[i].selected = !domains[currentIdx].ipEntries[i].selected
                    }
                }
            }
        }
    }
}