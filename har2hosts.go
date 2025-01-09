package main

import (
    "encoding/json"
    "fmt"
    "net/url"
    "os"
    "sort"
    "strings"
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

// DomainItem represents a domain with its selection state
type DomainItem struct {
    domain    string
    selected  bool
}

func drawList(s tcell.Screen, domains []DomainItem, currentIdx int) {
    s.Clear()
    _, height := s.Size()
    style := tcell.StyleDefault
    selectedStyle := style.Background(tcell.ColorGray)

    // Draw title
    title := "Space to toggle selection, Enter to save, Esc to quit"
    for i, r := range title {
        s.SetContent(i, 0, r, nil, style)
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
        prefix := "[ ]"
        if domains[i].selected {
            prefix = "[✓]"
        }
        
        lineStyle := style
        if i == currentIdx {
            lineStyle = selectedStyle
        }

        line := fmt.Sprintf("%s %s", prefix, domains[i].domain)
        for x, r := range line {
            s.SetContent(x, y, r, nil, lineStyle)
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
        fmt.Println("Usage: har2hosts <harfile>")
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

    // Extract unique domains
    domainsMap := make(map[string]bool)
    for _, entry := range har.Log.Entries {
        if parsedURL, err := url.Parse(entry.Request.URL); err == nil {
            hostname := parsedURL.Hostname()
            if hostname != "" {
                domainsMap[hostname] = true
            }
        }
    }

    // Convert to sorted slice of DomainItems
    var domains []DomainItem
    for domain := range domainsMap {
        domains = append(domains, DomainItem{domain: domain, selected: true})
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
                for _, item := range domains {
                    if item.selected {
                        output.WriteString(fmt.Sprintf("0.0.0.0 %s\n", item.domain))
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
                    domains[currentIdx].selected = !domains[currentIdx].selected
                }
            }
        }
    }
}