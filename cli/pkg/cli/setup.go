package cli

import (
	"fmt"

	"github.com/cline/cli/pkg/cli/setup"
	"github.com/spf13/cobra"
)

// NewSetupCommand creates the setup command
func NewSetupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Interactive setup wizard for API providers",
		Long: `Launch an interactive setup wizard to configure API providers for Cline CLI.

This wizard will guide you through:
- Selecting from 35+ supported AI providers
- Configuring API keys and settings
- Choosing models and capabilities
- Testing provider connections
- Setting default providers

All API keys are encrypted and stored securely in your Documents folder.

Examples:
  cline setup                    # Run the interactive setup wizard
  cline setup --help            # Show this help message`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetupWizard()
		},
	}

	return cmd
}

// runSetupWizard runs the interactive setup wizard
func runSetupWizard() error {
	wizard, err := setup.NewSetupWizard()
	if err != nil {
		return fmt.Errorf("failed to initialize setup wizard: %w", err)
	}

	return wizard.Run()
}
