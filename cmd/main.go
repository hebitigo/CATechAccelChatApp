package main

import (
	"context"

	"github.com/hebitigo/CATechAccelChatApp/db"
	"github.com/hebitigo/CATechAccelChatApp/router"

	_ "github.com/uptrace/bun/driver/pgdriver"
)

func main() {

	ctx := context.Background()
	db := db.GetDBConnection(ctx)
	defer db.Close()
	r := router.InitRouter(db, ctx)
	r.Run(":8080")
}
