package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/pgxpool"
)

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*.tmpl.html")

	router.GET("/diag/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"overall": "ok",
		})
	})

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl.html", gin.H{
			"title": "Invisible Identity",
		})
	})

	router.GET("/sign-up", func(c *gin.Context) {
		c.HTML(http.StatusOK, "sign-up.tmpl.html", gin.H{})
	})

	router.POST("/sign-up", func(c *gin.Context) {
		email := c.PostForm("emailAddress")
		passwordFirst := c.PostForm("passwordFirst")
		passwordAgain := c.PostForm("passwordAgain")

		log.Default().Printf("Sign up params: emailAddress: %s, passwordFirst: %s, passwordAgain: %s", email, passwordFirst, passwordAgain)

		// Todo: change to https://github.com/jackc/pgx
		pgConnectUrl := "postgres://invisibleprogrammer:invisiblepassword@localhost:5432/invisible-identity-db"

		connConfig := pgx.ConnConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "invisibleprogrammer",
			Password: "invisiblepassword",
			Database: "invisible-identity-db",
		}
		conn, err := pgx.Connect(connConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()

		var greeting string
		err = conn.QueryRow("select 'Hello, world!'").Scan(&greeting)
		if err != nil {
			fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
			os.Exit(1)
		}

		log.Default().Println(greeting)

		dbPool, err := pgxpool.Connect(context.Background(), pgConnectUrl)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Connect to pool failed: %v\n", err)
			os.Exit(1)
		}

		var userid int64
		resultSet := dbPool.QueryRow(context.Background(), "select userid from users where emailaddress = $1", email)

		err = resultSet.Scan(&userid)
		if err != nil && err.Error() != pgx.ErrNoRows.Error() {
			fmt.Fprintf(os.Stderr, "QueryRow from pool failed 1: %v\n", err)
			os.Exit(1)
		}

		// Todo: if there is no user found, register the user. If the user is already registered, give an error

		log.Default().Println(userid)
	})

	router.Run()
}
