package main

import (
	"flag"
	"fmt"
	"homenet/internal/config"
	"homenet/internal/dns"
	"homenet/internal/models"
	"homenet/internal/scanner"
	"homenet/internal/wol"
	"log"
	"os"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	baseStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)
		
	subTitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A49FA5")).
		MarginLeft(1)

	onlineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")) // Green
	offlineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")) // Grey
)

// Model stores the application state
type model struct {
	scanner   *scanner.Scanner
	dnsServer *dns.Server
	table     table.Model
	spinner   spinner.Model
	devices   []models.Device
	err       error
	scanning  bool
	subTitle  string
	alert     string
	// Stats
	totalQueries   uint64
	blockedQueries uint64
}

type tickMsg time.Time
type scanResultMsg []models.Device
type alertMsg string

// Init is the first function that runs
func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tickCmd(),
		scanCmd(m.scanner),
		waitForAlert(m.scanner),
	)
}

// Update handles messages and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			// Save on exit
			m.scanner.SaveDevices()
			return m, tea.Quit
		}
	
	case tickMsg:
		// Update Stats
		if m.dnsServer != nil {
			m.totalQueries, m.blockedQueries = m.dnsServer.GetStats()
		}
		return m, tea.Batch(tickCmd(), scanCmd(m.scanner))

	case scanResultMsg:
		m.devices = msg
		m.scanning = false
		m.updateTable()
		m.subTitle = fmt.Sprintf("Last updated: %s", time.Now().Format("15:04:05"))
		return m, nil
	
	case alertMsg:
		m.alert = string(msg)
		// Clear alert after 5 seconds
		go func() {
			time.Sleep(5 * time.Second)
		}()
		return m, waitForAlert(m.scanner)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View renders the UI
func (m model) View() string {
	title := `
   _____  ______  _   _  _______  _____  _   _  ______  _
  / ____||  ____|| \ | ||__   __||_   _|| \ | ||  ____|| |
 | (___  | |__   |  \| |   | |     | |  |  \| || |__   | |
  \___ \ |  __|  | .   |   | |     | |  | .   ||  __|  | |
  ____) || |____ | |\  |   | |    _| |_ | |\  || |____ | |____
 |_____/ |______||_| \_|   |_|   |_____||_| \_||______||______|
`
	header := lipgloss.JoinVertical(lipgloss.Center, 
		lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Render(title),
		subTitleStyle.Render(m.subTitle),
	)
	
	// Stats Line
	stats := fmt.Sprintf("\n %s Scanning... | Devices: %d | DNS Queries: %d | Blocked: %d", 
		m.spinner.View(), len(m.devices), m.totalQueries, m.blockedQueries)
	
	// Alert Banner
	if m.alert != "" {
		alertBanner := lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.Color("196")). // Red
			Foreground(lipgloss.Color("255")). // White
			Padding(0, 2).
			Render(" ⚠️  " + m.alert + "  ")
		stats = lipgloss.JoinVertical(lipgloss.Left, stats, "\n"+alertBanner)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		baseStyle.Render(m.table.View()),
		stats,
		"\n Press 'q' to quit.",
	)
}

// Helper commands
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func scanCmd(s *scanner.Scanner) tea.Cmd {
	return func() tea.Msg {
		return scanResultMsg(s.GetDevices())
	}
}

func waitForAlert(s *scanner.Scanner) tea.Cmd {
	return func() tea.Msg {
		return alertMsg(<-s.AlertChan)
	}
}
func (m *model) updateTable() {

	columns := []table.Column{

		{Title: "Status", Width: 10},

		{Title: "IP Address", Width: 16},

		{Title: "MAC / Manuf", Width: 25},

		{Title: "Hostname / Ports", Width: 30},

	}



	rows := []table.Row{}

	

	// Sort devices: Online first, then by IP

	sort.Slice(m.devices, func(i, j int) bool {

		if m.devices[i].IsOnline != m.devices[j].IsOnline {

			return m.devices[i].IsOnline // true (Online) comes before false (Offline)

		}

		return m.devices[i].IP < m.devices[j].IP

	})



	for _, d := range m.devices {

		status := "○ OFFLINE"

		if d.IsOnline {

			status = onlineStyle.Render("● ONLINE")

		} else {

			status = offlineStyle.Render("○ OFFLINE")

		}

		

		mac := d.MAC

		if mac == "" {

			mac = "N/A"

		} else if d.Manufacturer != "" {

			// Show Manufacturer if known, e.g. "Apple (00:1A...)"

			// Or just "Apple" to save space? Let's stack them or combine.

			mac = fmt.Sprintf("%s (%s)", d.Manufacturer, mac)

		}

		

		// Combine Hostname and Ports

		info := d.Hostname

		if len(d.Ports) > 0 {

			info = fmt.Sprintf("%s [%s]", info, d.Ports[0])

			if len(d.Ports) > 1 {

				info += fmt.Sprintf(" +%d", len(d.Ports)-1)

			}

		}



		rows = append(rows, table.Row{status, d.IP, mac, info})

	}



	m.table.SetColumns(columns)

	m.table.SetRows(rows)

}



func main() {
	wakePtr := flag.String("wake", "", "MAC address to wake (e.g., aa:bb:cc:dd:ee:ff)")
	configPtr := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	// Wake Mode
	if *wakePtr != "" {
		err := wol.Wake(*wakePtr)
		if err != nil {
			fmt.Printf("Error waking device: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Magic Packet sent.")
		return
	}

	// Load Configuration
	cfg, err := config.LoadConfig(*configPtr)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Setup Logging
	logFile, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		// Proceed without file logging if it fails? Or exit?
		// Let's just warn and proceed with stderr if we can't open file, 
		// but since TUI takes stderr/stdout, logs will be lost without file.
	} else {
		defer logFile.Close()
		log.SetOutput(logFile)
	}

	log.Println("Home Network Sentinel Starting...")

	// TUI Mode
	// Use configured subnet
	scanner := scanner.NewScanner(cfg.Subnet, cfg.DevicesFile)
	scanner.Start(5 * time.Second) // Background scan

	// Start DNS Server
	dnsServer := dns.NewServer(cfg.UpstreamDNS, cfg.BlockList)
	go func() {
		// Uses configured port
		dnsServer.Start(cfg.DNSPort)
	}()

	// Initialize Table
	t := table.New(
		table.WithColumns(nil),
		table.WithRows(nil),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	// Initialize Spinner
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	m := model{
		scanner:   scanner,
		dnsServer: dnsServer,
		table:     t,
		spinner:   sp,
		subTitle:  "Initializing...",
		alert:     "",
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}