package auth

import (
	"context"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/cline/cli/pkg/cli/global"
	"github.com/cline/grpc-go/cline"
)

// AuthAction represents the type of authentication action
type AuthAction string

const (
	AuthActionClineLogin    AuthAction = "cline_login"
	AuthActionProviderSetup AuthAction = "provider_setup"
)

// ShowAuthMenu displays the main auth menu and returns the selected action
func ShowAuthMenu() (AuthAction, error) {
	options := []string{
		"Authenticate with Cline account",
		"Configure API provider",
	}

	var choice string
	prompt := &survey.Select{
		Message: "What would you like to do?",
		Options: options,
	}

	if err := survey.AskOne(prompt, &choice); err != nil {
		return "", fmt.Errorf("failed to get menu choice: %w", err)
	}

	switch choice {
	case options[0]:
		return AuthActionClineLogin, nil
	case options[1]:
		return AuthActionProviderSetup, nil
	default:
		return "", fmt.Errorf("invalid choice")
	}
}

// HandleClineAuth handles Cline account authentication
func HandleClineAuth(ctx context.Context) error {
	fmt.Println("Authenticating with Cline...")
	
	// Check if already authenticated
	if isAuthenticated(ctx) {
		fmt.Println("You are already signed in to Cline.")
		
		// Ask if they want to sign out
		signOut := false
		prompt := &survey.Confirm{
			Message: "Would you like to sign out?",
			Default: false,
		}
		if err := survey.AskOne(prompt, &signOut); err != nil {
			return fmt.Errorf("failed to get sign out choice: %w", err)
		}
		
		if signOut {
			return signOutCline(ctx)
		}
		return nil
	}

	// Perform sign in
	if err := signInCline(ctx); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	fmt.Println("You are signed in!")
	return nil
}

// HandleProviderSetup launches the provider configuration wizard
func HandleProviderSetup() error {
	wizard, err := NewProviderWizard()
	if err != nil {
		return fmt.Errorf("failed to create provider wizard: %w", err)
	}

	return wizard.Run()
}

// isAuthenticated checks if the user is authenticated with Cline
func isAuthenticated(ctx context.Context) bool {
	client, err := global.GetDefaultClient(ctx)
	if err != nil {
		return false
	}

	_, err = client.Account.GetUserCredits(ctx, &cline.EmptyRequest{})
	return err == nil
}

// signInCline performs Cline account sign in
func signInCline(ctx context.Context) error {
	client, err := global.GetDefaultClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	_, err = client.Account.AccountLoginClicked(ctx, &cline.EmptyRequest{})
	if err != nil {
		return fmt.Errorf("failed to initiate login: %w", err)
	}

	return nil
}

// signOutCline performs Cline account sign out
func signOutCline(ctx context.Context) error {
	client, err := global.GetDefaultClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	if _, err = client.Account.AccountLogoutClicked(ctx, &cline.EmptyRequest{}); err != nil {
		return fmt.Errorf("failed to sign out: %w", err)
	}

	fmt.Println("You have been signed out of Cline.")
	return nil
}
