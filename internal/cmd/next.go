package cmd

import (
	"github.com/lucasassuncao/gopaper/internal/history"
	"github.com/spf13/cobra"
)

// NextCmd sets the next (more recent) wallpaper from history.
func NextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "next",
		Short: "Set the next wallpaper from history (after using prev)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return navigateHistory(history.Next, history.ErrAlreadyNewest)
		},
	}
}
