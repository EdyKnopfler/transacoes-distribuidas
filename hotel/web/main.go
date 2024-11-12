package main

import (
	"fmt"
	"net/http"

	"com.derso.aprendendo/conexoes"
	hotel "com.derso.aprendendo/hotel/negocio"
	"com.derso.aprendendo/sessoes"
	"com.derso.aprendendo/web"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BaseRequest struct {
	IdUsuario string `json:"idUsuario"`
	IdSessao  string `json:"idSessao"`
	IdVaga    string `json:"idVaga"`
}

var (
	gormPostgres *gorm.DB
	redisSessoes *conexoes.RedisConnection
	redisLock    *conexoes.RedisConnection
)

func alterarVaga(c *gin.Context, fnAlteracao func(request BaseRequest) error) {
	var request BaseRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "erro", "erro": "Requisição inválida"})
		return
	}

	err := sessoes.ExecutarSobBloqueio(request.IdVaga, redisLock, func() error {
		return fnAlteracao(request)
	})

	if err != nil {
		fmt.Println("Erro ao fazer pré-reserva", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "erro", "erro": "Requisição inválida"})
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func main() {
	db, err := conexoes.ConectarPostgreSQL("hotel")

	if err != nil {
		panic("Não foi possível conectar-se ao PostgreSQL.")
	}

	gormPostgres = db

	defer func() {
		sqlDB, _ := gormPostgres.DB()
		sqlDB.Close()
	}()

	gormPostgres.AutoMigrate(&hotel.Vaga{})

	redisSessoes = conexoes.ConectarRedis(conexoes.SESSION_DATABASE)
	defer redisSessoes.Fechar()

	redisLock = conexoes.ConectarRedis(conexoes.LOCK_DATABASE)
	defer redisLock.Fechar()

	srv, router := web.CreateServer(":8080")

	router.POST("/criar-sessao", func(c *gin.Context) {
		web.CriarSessao(redisSessoes, c)
	})

	router.POST("/marcar-vaga", func(c *gin.Context) {
		alterarVaga(c, func(request BaseRequest) error {
			return hotel.PreReserva(gormPostgres, request.IdVaga, request.IdUsuario)
		})
	})

	router.POST("/desmarcar-vaga", func(c *gin.Context) {
		alterarVaga(c, func(request BaseRequest) error {
			return hotel.Liberar(gormPostgres, request.IdVaga, request.IdUsuario)
		})
	})

	web.RunServer(srv)
}
