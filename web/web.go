package web

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"com.derso.aprendendo/conexoes"
	"com.derso.aprendendo/sessoes"
	"github.com/gin-gonic/gin"
)

type BaseRequest struct {
	IdUsuario string `json:"idUsuario"`
	IdSessao  string `json:"idSessao"`
	IdVaga    string `json:"idVaga"`
	IdAssento string `json:"idAssento"`
}

func CreateServer(hostAndPort string) (*http.Server, *gin.Engine) {
	router := gin.Default()

	srv := &http.Server{
		Addr:    hostAndPort,
		Handler: router,
	}

	return srv, router
}

func RunServer(server *http.Server) {
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("Erro ao criar servidor")
			panic(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	fmt.Println("Parando...")

	if err := server.Shutdown(ctx); err != nil {
		fmt.Println("Erro ao encerrar servidor:", err)
	}

	<-ctx.Done()
	fmt.Println("Servidor encerrado")
}

func CriarSessao(redis *conexoes.RedisConnection, c *gin.Context) {
	var request BaseRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "erro", "erro": "Requisição inválida"})
		return
	}

	idSessao, err := sessoes.CriarSessao(redis, request.IdUsuario)

	if err != nil {
		fmt.Println("Erro ao criar sessão:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao criar sessão"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"idSessao": idSessao})
}
