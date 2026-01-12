package commands

import (
	"fmt"

	"secretKeeper/internal/client/app" // Проверьте правильность пути импорта

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	v           = viper.New()
	application *app.App
)

func init() {
	rootCmd.PersistentFlags().StringP("server_addr", "s", "", "gRPC server address")
	rootCmd.PersistentFlags().StringP("cert", "c", "", "Path to TLS certificate file")

	viper.BindPFlag("CLIENT_SERVER_ADDR", rootCmd.PersistentFlags().Lookup("server_addr"))
	viper.BindPFlag("CLIENT_CERT_PATH", rootCmd.PersistentFlags().Lookup("cert"))
}

var rootCmd = &cobra.Command{
	Use:   "secretKeeper",
	Short: "Приложение для безопасного хранения секретов",
	Long:  `SecretKeeper - это безопасное CLI приложение для управления паролями и файлами.`,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		application, err = app.NewApp(v)
		if err != nil {
			return fmt.Errorf("ошибка инициализации приложения: %w", err)
		}
		return nil
	},

	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if application != nil {
			return application.Close()
		}
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}
