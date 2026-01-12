package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// CLIInteractor реализует интерфейс UserInteractor для CLI
type CLIInteractor struct {
	reader *bufio.Reader
}

// NewCLIInteractor создает новый CLI интерактор
func NewCLIInteractor() *CLIInteractor {
	return &CLIInteractor{
		reader: bufio.NewReader(os.Stdin),
	}
}

// ConfirmCacheUsage запрашивает подтверждение у пользователя на использование кеша
func (i *CLIInteractor) ConfirmCacheUsage(secretID string) (bool, error) {
	fmt.Printf("\n⚠️  Сервер недоступен. Секрет '%s' найден в локальном кеше.\n", secretID)
	fmt.Print("Использовать кешированную версию секрета? (y/n): ")

	response, err := i.reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("ошибка чтения ввода: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes" || response == "да", nil
}

var getSecretCmd = &cobra.Command{
	Use:   "get [secret-id]",
	Short: "Получить секрет по ID",
	Long:  `Получает секрет с сервера. Если сервер недоступен, предлагает использовать кешированную версию.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		secretID := args[0]

		ctx := context.Background()

		// Устанавливаем интерактор для UserService
		interactor := NewCLIInteractor()
		application.UserService.SetInteractor(interactor)

		// Получаем секрет с поддержкой оффлайн режима
		secret, err := application.UserService.GetSecretWithOfflineFallback(ctx, secretID)
		if err != nil {
			return fmt.Errorf("не удалось получить секрет: %w", err)
		}

		fmt.Printf("\n✅ Секрет успешно получен!\n")
		fmt.Printf("ID: %s\n", secret.Meta.ID)
		fmt.Printf("Тип: %s\n", secret.Meta.Type)

		if name, ok := secret.Meta.Extra["name"]; ok {
			fmt.Printf("Название: %s\n", name)
		}

		if description, ok := secret.Meta.Extra["description"]; ok {
			fmt.Printf("Описание: %s\n", description)
		}

		// Выводим содержимое секрета
		fmt.Println("\nСодержимое:")
		data := make([]byte, 1024)
		n, err := secret.Data.Read(data)
		if err != nil && err.Error() != "EOF" {
			return fmt.Errorf("ошибка чтения данных секрета: %w", err)
		}
		fmt.Println(string(data[:n]))

		return nil
	},
}

var listSecretsCmd = &cobra.Command{
	Use:   "list",
	Short: "Получить список всех секретов",
	Long:  `Получает список всех секретов пользователя с сервера.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		secrets, err := application.UserService.GetList(ctx)
		if err != nil {
			return fmt.Errorf("не удалось получить список секретов: %w", err)
		}

		if len(secrets) == 0 {
			fmt.Println("У вас пока нет сохраненных секретов.")
			return nil
		}

		fmt.Printf("\n📋 Найдено секретов: %d\n\n", len(secrets))
		for i, secret := range secrets {
			fmt.Printf("%d. ID: %s\n", i+1, secret.ID)
			fmt.Printf("   Тип: %s\n", secret.Type)

			if name, ok := secret.Extra["name"]; ok {
				fmt.Printf("   Название: %s\n", name)
			}

			if description, ok := secret.Extra["description"]; ok {
				fmt.Printf("   Описание: %s\n", description)
			}

			fmt.Println()
		}

		return nil
	},
}

var deleteSecretCmd = &cobra.Command{
	Use:   "delete [secret-id]",
	Short: "Удалить секрет по ID",
	Long:  `Удаляет секрет с сервера по указанному ID.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		secretID := args[0]

		ctx := context.Background()

		status, err := application.UserService.DeleteSecret(ctx, secretID)
		if err != nil {
			return fmt.Errorf("не удалось удалить секрет: %w", err)
		}

		fmt.Printf("\n✅ Секрет успешно удален! Статус: %s\n", status)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getSecretCmd)
	rootCmd.AddCommand(listSecretsCmd)
	rootCmd.AddCommand(deleteSecretCmd)
}
