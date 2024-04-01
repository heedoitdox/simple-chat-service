package network

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
)

type Network struct {
	engin *gin.Engine
}

func NewServer() *Network {
	n := &Network{engin: gin.New()}

	// Use() 모든 API 나 라우터를 통해 범용적인 처리
	// gin 프레임워크를 보면 아래 두개는 기본적으로 많이들 사용한다.
	n.engin.Use(gin.Logger())   // API 가 들어오는 것에 로그 찍기
	n.engin.Use(gin.Recovery()) // panic 으로 인해 서버가 죽었을 때 자동으로 서버를 올려주는 것
	n.engin.Use(cors.New(cors.Config{
		AllowWebSockets:  true,
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))

	r := NewRoom()
	go r.RunInit() // 고루틴 -> 백그라운드에서 동작을 해라 라는 의미

	n.engin.GET("/room", r.SocketServe)

	return n
}

// 서버를 시작할 수 있는 함수
func (n *Network) StartServer() error {
	log.Println("Starting server........")
	return n.engin.Run(":8080")
}
