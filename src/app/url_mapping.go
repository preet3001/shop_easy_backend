package app

import (
	"github.com/preet3001/shop_easy_backend/src/controllers/ping_controller"
)

func mapUrls(){
	router.GET("/ping",pingcontroller.Ping)
}