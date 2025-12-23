package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"secretKeeper/internal/client/transport"
	pb "secretKeeper/internal/server/proto"
	"strings"
)

// Функция точки входа клиентского приложения (CLI)
func main() {
	if err := run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// Функция основного цикла работы CLI приложения
func run() error {
	ctx := context.Background()
	client, err := transport.SetupClient(ctx)
	if err != nil {
		return fmt.Errorf("setup failed: %w", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("SecretKeeper Client Started")
	fmt.Println("Commands: register <user> <pass>, login <user> <pass>, file <name> <path>, help, quit")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		cmd := parts[0]
		switch cmd {
		case "quit", "exit":
			return nil
		case "help":
			fmt.Println("Available commands:")
			fmt.Println("  register <username> <password>")
			fmt.Println("  login    <username> <password>")
			fmt.Println("  file     <secret_name> <file_path>")
			fmt.Println("  quit")
		case "register":
			if len(parts) != 3 {
				fmt.Println("Usage: register <username> <password>")
				continue
			}
			if err := client.Register(ctx, parts[1], parts[2]); err != nil {
				fmt.Printf("Register failed: %v\n", err)
			} else {
				fmt.Println("Registration successful!")
			}
		case "login":
			if len(parts) != 3 {
				fmt.Println("Usage: login <username> <password>")
				continue
			}
			if err := client.Login(ctx, parts[1], parts[2]); err != nil {
				fmt.Printf("Login failed: %v\n", err)
			} else {
				fmt.Println("Login successful!")
			}
		case "file":
			if len(parts) != 3 {
				fmt.Println("Usage: file <secret_name> <file_path>")
				continue
			}
			if !client.IsAuthorized() {
				fmt.Println("Error: You must login first")
				continue
			}

			name := parts[1]
			path := parts[2]

			data, err := os.ReadFile(path)
			if err != nil {
				fmt.Printf("Failed to read file: %v\n", err)
				continue
			}

			meta := &pb.SecretMetadata{
				Name: name,
				Type: pb.SecretType_SECRET_TYPE_FILE,
				PublicMeta: map[string]string{
					"filename": path,
				},
			}

			resp, err := client.CreateSecret(ctx, meta, data)
			if err != nil {
				fmt.Printf("Upload failed: %v\n", err)
			} else {
				fmt.Printf("Success! Secret ID: %s (status: %s)\n", resp.Id, resp.Status)
			}

		default:
			fmt.Println("Unknown command. Type 'help'")
		}
	}
	return nil
}
