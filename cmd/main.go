package main

import (
	"CATechAccelChatApp/repository"
	"CATechAccelChatApp/router"
	"context"

	_ "github.com/uptrace/bun/driver/pgdriver"
)

func main() {

	ctx := context.Background()
	db := repository.GetDBConnection(ctx)
	defer db.Close()
	r := router.InitRouter(db, ctx)
	r.Run(":8080")
}
