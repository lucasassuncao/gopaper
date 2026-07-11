package cmd

import (
	"fmt"

	"github.com/lucasassuncao/gopaper/internal/helper"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// MonitorsCmd lists the connected monitors with the 1-based index gopaper
// uses for configuration.behavior.monitor ("monitor1", "monitor2", ...) and
// categories[].monitor, so users can tell which physical monitor a given
// index refers to.
func MonitorsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "monitors",
		Short: "List connected monitors and the index gopaper uses for each",
		Long: `List every monitor gopaper can target, with the 1-based index used by
configuration.behavior.monitor ("monitor1", "monitor2", ...) and
categories[].monitor.

Name is the monitor's EDID-reported name (e.g. "ASUS VG32VQ1B"), read via
WMI; falls back to the 3-letter manufacturer code (e.g. "BOE") when a
monitor's EDID doesn't set a friendly name, which is common for laptop
panels. Position is each monitor's desktop rectangle, in the same
arrangement as Windows Display Settings: the monitor at 0,0 is the
reference point, and others are offset from it (e.g. a monitor at
"1920,0" sits to its right).`,
		Example: `  gopaper monitors`,
		RunE: func(cmd *cobra.Command, args []string) error {
			details, err := helper.ListMonitorDetails()
			if err != nil {
				return fmt.Errorf("could not list monitors: %w", err)
			}

			table := pterm.TableData{{"Index", "Name", "Position", "Size", "Device Path"}}
			for _, d := range details {
				name := d.Name
				if name == "" {
					name = "-"
				}
				table = append(table, []string{
					fmt.Sprintf("monitor%d", d.Index),
					name,
					fmt.Sprintf("%d,%d", d.Left, d.Top),
					fmt.Sprintf("%dx%d", d.Right-d.Left, d.Bottom-d.Top),
					d.DevicePath,
				})
			}

			return pterm.DefaultTable.WithHasHeader().WithData(table).Render()
		},
	}
}
