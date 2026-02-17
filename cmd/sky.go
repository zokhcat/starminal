package cmd

import (
	"os"

	"starminal/src"
	"starminal/utils"

	tea "github.com/charmbracelet/bubbletea"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var skyCmd = &cobra.Command{
	Use:   "sky",
	Short: "Render the stars in a sky",
	Run: func(cmd *cobra.Command, args []string) {
		catalog := utils.LoadCatalog("data/hyg_v42.csv")

		model := src.NewSkyModel(catalog)
		p := tea.NewProgram(model, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			log.Error(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(skyCmd)
}
