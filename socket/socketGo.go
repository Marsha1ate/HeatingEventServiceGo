package socket

import (
	socketio "github.com/googollee/go-socket.io"
)

// Server будет хранить глобальную ссылку на экземпляр сервера Socket.IO
var Server *socketio.Server