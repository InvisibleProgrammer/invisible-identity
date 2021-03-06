package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

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

		if passwordFirst != passwordAgain {
			c.HTML(http.StatusBadRequest, "sign-up.tmpl.html", gin.H{})
		}

		passwordHash, err := hashPassword(passwordFirst)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to create password hash: %v\n", err)
		}

		pgConnectUrl := "postgres://invisibleprogrammer:invisiblepassword@localhost:5432/invisible-identity-db"

		dbPool, err := pgxpool.Connect(context.Background(), pgConnectUrl)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Connect to pool failed: %v\n", err)
			os.Exit(1)
		}

		var userid int64
		resultSet := dbPool.QueryRow(context.Background(), "select userid from users where emailaddress = $1", email)

		err = resultSet.Scan(&userid)
		if err != nil && err.Error() != pgx.ErrNoRows.Error() {
			fmt.Fprintf(os.Stderr, "User %v is already registered.", email)
			os.Exit(1)
		}

		if userid > 0 {
			return
		}

		now := time.Now()
		resultSet = dbPool.QueryRow(context.Background(), "insert into Users (EmailAddress, Activated, RecordedAt, UpdatedAt) values ($1, $2, $3, $3) returning UserId", email, false, now)

		err = resultSet.Scan(&userid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Connect to pool failed: %v\n", err)
			os.Exit(1)
		}

		log.Default().Printf("%v is registered with id %v", email, userid)

		commandTag, err := dbPool.Exec(context.Background(), "insert into PasswordHashes(UserId, PasswordHash, RecordedAt, UpdatedAt) values ($1, $2, $3, $3);", userid, passwordHash, now)
		if err != nil || commandTag.RowsAffected() != 1 {
			fmt.Fprintf(os.Stderr, "Cannot store password hash for user %v\n", userid)
		}

		// Todo: generate activation ticket: https://gist.github.com/dopey/c69559607800d2f2f90b1b1ed4e550fb

		c.Redirect(http.StatusMovedPermanently, "/")

	})

	router.Run()
}
