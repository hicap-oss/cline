package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cline/cli/pkg/cli/auth"
	"github.com/cline/cli/pkg/cli/global"
	"github.com/cline/grpc-go/cline"
	"github.com/spf13/cobra"
)

var isSessionAuthenticated bool

func NewAuthCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "auth [provider] [key]",
		Short: "Authenticate with Cline or configure API providers",
		Long: `Authenticate with Cline account or configure API providers.

Usage modes:
  cline auth                     # Interactive menu: choose Cline auth or provider setup
  cline auth [provider]          # Configure specific provider (prompts for API key)
  cline auth [provider] [key]    # Fast setup with provider and API key

Examples:
  cline auth                          # Show interactive menu
  cline auth anthropic                # Configure Anthropic (will prompt for key)
  cline auth anthropic sk-ant-xxx     # Fast setup with Anthropic
  cline auth openrouter sk-or-xxx     # Fast setup with OpenRouter`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleAuthCommand(cmd.Context(), args)
		},
	}
}

func handleAuthCommand(ctx context.Context, args []string) error {
	// Route based on number of arguments
	switch len(args) {
	case 0:
		// No args: Show menu to choose between Cline auth or provider setup
		return handleAuthMenu(ctx)
	case 1:
		// One arg: Provider ID only, prompt for API key
		return auth.FastSetup(args[0], "")
	case 2:
		// Two args: Provider ID and API key
		return auth.FastSetup(args[0], args[1])
	default:
		return fmt.Errorf("too many arguments. Usage: cline auth [provider] [key]")
	}
}

func handleAuthMenu(ctx context.Context) error {
	// Show menu to choose between Cline auth and provider setup
	action, err := auth.ShowAuthMenu()
	if err != nil {
		return err
	}

	switch action {
	case auth.AuthActionClineLogin:
		return auth.HandleClineAuth(ctx)
	case auth.AuthActionProviderSetup:
		return auth.HandleProviderSetup()
	default:
		return fmt.Errorf("invalid action")
	}
}

func signOut(ctx context.Context) error {
	client, err := global.GetDefaultClient(ctx)
	if err != nil {
		return err
	}

	if _, err = client.Account.AccountLogoutClicked(ctx, &cline.EmptyRequest{}); err != nil {
		return err
	}

	isSessionAuthenticated = false
	fmt.Println("You have been signed out of Cline.")
	return nil
}

func signOutDialog(ctx context.Context) error {
	fmt.Print("You are already signed in to Cline.\nWould you like to sign out? (y/N): ")

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return nil
	}

	response := strings.ToLower(strings.TrimSpace(scanner.Text()))
	if response == "y" || response == "yes" {
		if err := signOut(ctx); err != nil {
			fmt.Printf("Failed to sign out: %v\n", err)
			return err
		}
	}
	return nil
}

func signIn(ctx context.Context) error {
	if IsAuthenticated(ctx) {
		return nil
	}

	verboseLog("Ensuring default instance exists...")
	if err := ensureDefaultInstance(ctx); err != nil {
		verboseLog("Failed to ensure default instance: %v", err)
		return err
	}

	verboseLog("Default instance ensured successfully.")
	time.Sleep(2 * time.Second) // Allow services to start

	client, err := global.GetDefaultClient(ctx)
	if err != nil {
		verboseLog("Failed to obtain client: %v", err)
		return err
	}

	_, err = client.Account.AccountLoginClicked(ctx, &cline.EmptyRequest{})
	if err != nil {
		verboseLog("Failed to login: %v", err)
		return err
	}

	isSessionAuthenticated = true
	verboseLog("Login successful")
	return nil
}

func IsAuthenticated(ctx context.Context) bool {
	if isSessionAuthenticated {
		return true
	}

	client, err := global.GetDefaultClient(ctx)
	if err != nil {
		return false
	}

	_, err = client.Account.GetUserCredits(ctx, &cline.EmptyRequest{})
	return err == nil
}

func verboseLog(format string, args ...interface{}) {
	if global.Config != nil && global.Config.Verbose {
		fmt.Printf("[VERBOSE] "+format+"\n", args...)
	}
}
