package Forum

import "net/http"


type config struct {
	user     user
	server server
}
type server struct {
	Mode         string
	Addr         string
	ReadTimeout  int
	WriteTimeout int
}
type user struct {
    User         string
    Email        string
	Password     string
	Host         string
	Name         string
}


func Forum() {
	
}

func HandlerForum(w http.ResponseWriter, r *http.Request) {
}