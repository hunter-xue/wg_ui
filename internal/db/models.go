package db

type Server struct {
	ID          int64
	Name        string
	Address     string
	ListenPort  int
	PrivateKey  string
	PublicKey   string
	MTU         int
	DNS         string
	PostUp      string
	PostDown    string
	Description string // used to store endpoint (public IP)
	Comments    string
}

type Client struct {
	ID                  int64
	ServerID            int64
	Name                string
	Address             string
	ListenPort          int
	PrivateKey          string
	PublicKey           string
	AllowIPs            string
	MTU                 int
	DNS                 string
	Description         string // stores generated client config
	Comments            string
	Disabled            int
	PersistentKeepalive int
}
