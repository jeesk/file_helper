package main

import "file-helper/db"

func main() {
	sqlite := db.MsgbackendSqlite{}
	sqlite.InitSqlite("/home/song/fp.db")
	startServer(&sqlite)
	select {}
}
