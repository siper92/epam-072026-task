package db

type User struct {
	ID           string
	Username     string
	PasswordHash string
	Role         string
	CreatedAt    string
}

type Game struct {
	ID        string
	PlayerX   string
	PlayerO   string
	Grid      string
	Status    string
	WinnerID  string
	MoveCount int64
}
