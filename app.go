package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	upgrader = websocket.Upgrader{}
)

// This defines the structure of the messages that
// will pass back and forth to the server.
type GameMessage struct {
	Command string `json:"command"`
	Data    string `json:"data"`
}

type ServerInfo struct {
	Title    string `json:"title"`
	Contents string `json:"contents"`
}

var db = make(map[int]ServerInfo)

func gamesocket(c echo.Context) error {
	db[0] = ServerInfo{Title: "Record 1", Contents: "Something from the server about record 1"}
	db[1] = ServerInfo{Title: "Record 2", Contents: "Info from server about record 2"}
	// This function is mandatory to get the gorilla websockets library to work
	// in production, you would have a list of authorized origins that are allowed to
	// upgrade to a socket connection.
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	defer ws.Close()

	for {
		g := GameMessage{}
		err = ws.ReadJSON(&g)
		if err != nil {
			c.Logger().Error(err)
		}

		if g.Command == "View Record" {
			// cast to the Data to an integer for this case
			recordNum, err := strconv.Atoi(g.Data)
			if err != nil {
				c.Logger().Error(err)
				ws.WriteJSON(&GameMessage{Command: "Error", Data: "Could not transform record number to integer"})
			}

			msg := GameMessage{
				Command: "Send Record",
			}
			data, err := json.Marshal(db[recordNum])
			if err != nil {
				c.Logger().Error(err)
				ws.WriteJSON(&GameMessage{Command: "Error", Data: "Could not transform record data to string"})
			}
			msg.Data = string(data)
			ws.WriteJSON(&msg)
		}
	}
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	// Because we are running the server on a different host/port combo than the
	// frontend with the Unity3d embed, we need to add CORS middleware.
	e.Use(middleware.CORS())
	e.GET("/ws", gamesocket)
	e.Logger.Fatal(e.Start(":8000"))
}
