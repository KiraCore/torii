package configinternal

// Common - common settings for microservice (server options, socket port and etc)
type Common struct {
	HttpServer   `json:"http_server"`
	SocketServer `json:"socket_server"`
	WebSocket    `json:"web_socket"`
}

type WebSocket struct {
	Enabled bool   `json:"enabled"`
	Token   string `json:"token"`
	Url     string `json:"url"`
}

type HttpServer struct {
	Enabled bool   `json:"enabled"`
	Host    string `json:"host"`
	Port    string `json:"port"`
}

type SocketServer struct {
	Enabled bool   `json:"enabled"`
	Host    string `json:"host"`
	Port    string `json:"port"`
}
