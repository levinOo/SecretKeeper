package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(pingCmd)
}

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Проверяет доступность gRPC сервера",
	Run: func(cmd *cobra.Command, args []string) {

		if application == nil {
			fmt.Println("Ошибка: приложение не инициализировано")
			os.Exit(1)
		}

		err := application.Ping(context.Background())
		if err != nil {
			fmt.Printf("Пинг не удался: %v\n", err)
			return
		}

		fmt.Println("✅ Успешный пинг gRPC сервера.")
	},
}
