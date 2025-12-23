package domain

// Тип секрета (enum)
type SecretType int

// Константы типов секретов
const (
	SecretTypeUnknown SecretType = iota
	SecretTypeCredentials
	SecretTypeText
	SecretTypeCard
	SecretTypeFile
)

// Структура данных для логинов/паролей
type CredentialsPayload struct {
	Name        string "name"
	Login       string `json:"login"`
	Password    string `json:"password"`
	URL         string `json:"url,omitempty"`
	Description string `json:"description,omitempty"`
}

// Структура данных для банковских карт
type CardPayload struct {
	Name        string "name"
	Number      string `json:"number"` // "4242 4242..."
	HolderName  string `json:"holder_name"`
	ExpiryMonth int    `json:"expiry_month"`
	ExpiryYear  int    `json:"expiry_year"`
	CVV         string `json:"cvv"`
}

// Структура данных для простого текста
type TextPayload struct {
	Name string "name"
	Text string `json:"text"`
}

// Структура данных для файлов
type FilePayload struct {
	Name string "name"
}

// Структура метаданных секрета
type SecretMeta struct {
	Name       string            // Имя секрета ("Мой VK")
	Type       SecretType        // Тип (Credentials, Card...)
	PublicData map[string]string // Доп. инфа ("login": "ivan", "file_ext": ".pdf")
}
