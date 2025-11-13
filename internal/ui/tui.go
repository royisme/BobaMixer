package ui

import (
	"fmt"

	"github.com/vantagecraft-dev/bobamixer/internal/store/config"
)

// Model represents the TUI state
type Model struct {
	home          string
	activeProfile string
	profiles      config.Profiles
	todayStats    Stats
	width         int
	height        int
}

// Stats represents usage statistics
type Stats struct {
	TotalTokens int
	TotalCost   float64
	Sessions    int
}

// Run starts the TUI
func Run(home string) error {
	// Load profiles
	profiles, err := config.LoadProfiles(home)
	if err != nil {
		return fmt.Errorf("failed to load profiles: %w", err)
	}

	// Get active profile
	activeProfile := getActiveProfile(home)
	if activeProfile == "" && len(profiles) > 0 {
		// Default to first profile
		for k := range profiles {
			activeProfile = k
			break
		}
	}

	// For now, just print a simple TUI
	printSimpleTUI(home, activeProfile, profiles)

	return nil
}

func getActiveProfile(home string) string {
	// TODO: Read from active_profile file
	return ""
}

func printSimpleTUI(home, activeProfile string, profiles config.Profiles) {
	fmt.Println("╭─ BobaMixer ──────────────────────────────────────╮")
	fmt.Println("│                                                  │")

	if activeProfile != "" {
		prof := profiles[activeProfile]
		fmt.Printf("│ Active Profile: %-32s │\n", prof.Name)
		fmt.Printf("│ Model: %-41s │\n", prof.Model)
		fmt.Printf("│ Adapter: %-39s │\n", prof.Adapter)
	} else {
		fmt.Println("│ No active profile                               │")
	}

	fmt.Println("│                                                  │")
	fmt.Println("│ Today's Usage                                    │")
	fmt.Println("│   Tokens: 0                                      │")
	fmt.Println("│   Cost: $0.00                                    │")
	fmt.Println("│   Sessions: 0                                    │")
	fmt.Println("│                                                  │")
	fmt.Println("│ Available Profiles:                              │")

	count := 0
	for key, prof := range profiles {
		if count >= 3 {
			fmt.Println("│   ... and more                                   │")
			break
		}
		fmt.Printf("│   - %-44s │\n", key+" ("+prof.Model+")")
		count++
	}

	fmt.Println("│                                                  │")
	fmt.Println("│ Commands:                                        │")
	fmt.Println("│   boba use <profile>  - Switch profile           │")
	fmt.Println("│   boba stats --today  - Show today's stats       │")
	fmt.Println("│   boba doctor         - Check configuration      │")
	fmt.Println("│                                                  │")
	fmt.Println("╰──────────────────────────────────────────────────╯")
}
