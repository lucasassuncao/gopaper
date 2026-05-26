package cmd

import (
	"github.com/lucasassuncao/gopaper/internal/history"
	"github.com/spf13/cobra"
)

// PrevCmd sets the previous wallpaper from history.
func PrevCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "prev",
		Short: "Set the previous wallpaper from history",
		RunE: func(cmd *cobra.Command, args []string) error {
			return navigateHistory(history.Prev, history.ErrAlreadyOldest)
		},
	}
}
