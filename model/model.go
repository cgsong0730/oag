package model

type Auth struct {
	Hostname string
	Port     string
	Username string
	Password string
}

type OpenData struct {
	Params string
	Data   string
}
