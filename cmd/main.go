package main

import (
	"context"

	_ "github.com/uptrace/bun/driver/pgdriver"

	"github.com/hebitigo/CATechAccelChatApp/db"
	"github.com/hebitigo/CATechAccelChatApp/router"
)

func main() {

	ctx := context.Background()
	db := db.GetDBConnection(ctx)
	defer db.Close()
	r := router.InitRouter(db, ctx)
	r.Run(":8080")
}
