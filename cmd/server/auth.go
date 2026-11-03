package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"golang.org/x/term"

	"github.com/SailmanSeeulater/automorph/internal/auth"
	"github.com/SailmanSeeulater/automorph/internal/store"
)

// createAdmin creates an admin user, or resets the password if the email already exists.
//
//	automorph create-admin -email you@example.com [-name "Your Name"]
func createAdmin(ctx context.Context, q *store.Queries, args []string) error {
	fs := flag.NewFlagSet("create-admin", flag.ContinueOnError)
	email := fs.String("email", "", "admin email (required)")
	name := fs.String("name", "", "display name")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *email == "" {
		return errors.New(`usage: create-admin -email you@example.com [-name "Your Name"]`)
	}

	password, err := promptPassword()
	if err != nil {
		return err
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	user, err := q.UpsertAdmin(ctx, store.UpsertAdminParams{
		Email:        *email,
		PasswordHash: hash,
		DisplayName:  *name,
	})
	if err != nil {
		return err
	}
	fmt.Printf("admin ready: id=%d email=%s\n", user.ID, user.Email)
	return nil
}

// promptPassword reads the password twice without echoing it to the terminal.
func promptPassword() (string, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return "", errors.New("create-admin must be run in an interactive terminal")
	}
	fmt.Print("Password: ")
	first, err := term.ReadPassword(fd)
	fmt.Println()
	if err != nil {
		return "", err
	}
	fmt.Print("Repeat password: ")
	second, err := term.ReadPassword(fd)
	fmt.Println()
	if err != nil {
		return "", err
	}
	if string(first) != string(second) {
		return "", errors.New("passwords don't match")
	}
	if len(first) < 12 {
		return "", errors.New("use at least 12 characters")
	}
	return string(first), nil
}
