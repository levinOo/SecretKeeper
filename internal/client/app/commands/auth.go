package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(registerCmd)
}

var loginCmd = &cobra.Command{
	Use:  "login [username] [password]",
	Args: cobra.ExactArgs(2), // Требуем ровно 2 аргумента
	RunE: func(cmd *cobra.Command, args []string) error {
		username := args[0]
		password := args[1]

		ctx := context.Background()

		err := application.AuthService.Login(ctx, username, password)
		if err != nil {
			return err
		}

		fmt.Println("Успешный вход в систему!")
		return nil
	},
}

var registerCmd = &cobra.Command{
	Use:  "register [username] [password]",
	Args: cobra.ExactArgs(2), // Требуем ровно 2 аргумента
	RunE: func(cmd *cobra.Command, args []string) error {
		username := args[0]
		password := args[1]

		ctx := context.Background()

		err := application.AuthService.Register(ctx, username, password)
		if err != nil {
			return err
		}

		fmt.Println("Регистрация прошла успешно!")
		return nil
	},
}
