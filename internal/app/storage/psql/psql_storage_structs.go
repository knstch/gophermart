package psql

type Users struct {
	Login    string `bun:"login"`
	Password string `bun:"password"`
}
