module github.com/whitewhale1075/urmy_app

go 1.16

replace (
	github.com/whitewhale1075/urmy_handler v0.0.0 => /home/johnoh/goproject/urmy/urmy_handler
)

require (
	github.com/gorilla/mux v1.8.0
	github.com/unrolled/render v1.0.3
	github.com/urfave/negroni v1.0.0
	github.com/whitewhale1075/urmy_handler v0.0.2
)
