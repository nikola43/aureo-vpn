package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	loginEmail    string
	loginPassword string
)

// loginCmd handles user login
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to your Aureo VPN account",
	Long: `Login to your Aureo VPN account using email and password.

You can provide credentials via flags or interactively:
  aureo login
  aureo login --email user@example.com
  aureo login -e user@example.com -p yourpassword`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if already logged in
		if APIClient.IsAuthenticated() {
			email, _ := APIClient.GetCurrentUser()
			fmt.Printf("Already logged in as %s\n", email)
			fmt.Println("Use 'aureo logout' to sign out first")
			return nil
		}

		// Get email
		email := loginEmail
		if email == "" {
			email = promptInput("Email: ")
		}

		// Get password
		password := loginPassword
		if password == "" {
			password = promptPassword("Password: ")
		}

		fmt.Println("\nLogging in...")

		resp, err := APIClient.Login(email, password)
		if err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		fmt.Printf("\nWelcome back, %s!\n", resp.User.Username)
		fmt.Printf("Subscription: %s\n", resp.User.SubscriptionTier)
		fmt.Println("\nYou can now use 'aureo connect' to connect to a VPN node")
		return nil
	},
}

var (
	registerEmail    string
	registerPassword string
	registerUsername string
	registerFullName string
)

// registerCmd handles user registration
var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Create a new Aureo VPN account",
	Long: `Create a new Aureo VPN account.

You can provide details via flags or interactively:
  aureo register
  aureo register --email user@example.com --username myuser`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if already logged in
		if APIClient.IsAuthenticated() {
			email, _ := APIClient.GetCurrentUser()
			fmt.Printf("Already logged in as %s\n", email)
			fmt.Println("Use 'aureo logout' to sign out first")
			return nil
		}

		// Get email
		email := registerEmail
		if email == "" {
			email = promptInput("Email: ")
		}

		// Get username
		username := registerUsername
		if username == "" {
			username = promptInput("Username: ")
		}

		// Get full name
		fullName := registerFullName
		if fullName == "" {
			fullName = promptInput("Full Name: ")
		}

		// Get password
		password := registerPassword
		if password == "" {
			password = promptPassword("Password (min 8 characters): ")
			confirmPassword := promptPassword("Confirm Password: ")
			if password != confirmPassword {
				return fmt.Errorf("passwords do not match")
			}
		}

		fmt.Println("\nCreating account...")

		resp, err := APIClient.Register(email, password, username, fullName)
		if err != nil {
			return fmt.Errorf("registration failed: %w", err)
		}

		fmt.Printf("\nWelcome to Aureo VPN, %s!\n", resp.User.Username)
		fmt.Printf("Your account has been created with the %s subscription tier.\n", resp.User.SubscriptionTier)
		fmt.Println("\nYou are now logged in. Use 'aureo connect' to connect to a VPN node")
		return nil
	},
}

// logoutCmd handles user logout
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from your Aureo VPN account",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !APIClient.IsAuthenticated() {
			fmt.Println("Not logged in")
			return nil
		}

		email, _ := APIClient.GetCurrentUser()

		if err := APIClient.Logout(); err != nil {
			return fmt.Errorf("logout failed: %w", err)
		}

		fmt.Printf("Logged out from %s\n", email)
		return nil
	},
}

func init() {
	// Login flags
	loginCmd.Flags().StringVarP(&loginEmail, "email", "e", "", "Email address")
	loginCmd.Flags().StringVarP(&loginPassword, "password", "p", "", "Password")

	// Register flags
	registerCmd.Flags().StringVarP(&registerEmail, "email", "e", "", "Email address")
	registerCmd.Flags().StringVarP(&registerPassword, "password", "p", "", "Password")
	registerCmd.Flags().StringVarP(&registerUsername, "username", "u", "", "Username")
	registerCmd.Flags().StringVarP(&registerFullName, "name", "n", "", "Full name")
}

// promptInput reads a line of input from stdin
func promptInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// promptPassword reads a password without echoing
func promptPassword(prompt string) string {
	fmt.Print(prompt)
	password, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Print newline after password input
	if err != nil {
		return ""
	}
	return string(password)
}
