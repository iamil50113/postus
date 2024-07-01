package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"math/rand"
	"postus/internal/config"
	"postus/internal/repository/postgres"
	"postus/internal/service/comment"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func main() {
	cfg := config.MustLoad()

	log := &slog.Logger{}

	postgresPool, err := pgxpool.New(context.Background(), fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", cfg.PostgresConfig.Host, cfg.PostgresConfig.Port, cfg.PostgresConfig.User, cfg.PostgresConfig.Pass, cfg.PostgresConfig.DBName))
	if err != nil {
		panic("db connection error")
	}

	defer postgresPool.Close()

	st, err := postgres.New(postgresPool)
	if err != nil {
		panic("")
	}

	//postService := post.New(log, st, st, st)
	commentService := comment.New(log, cfg.PostusConfig.CommentLenLimit, cfg.PostusConfig.PaginationCommentsLimit, st, st, st, st)
	//userService := user.New(log, st)

	//for i := 0; i < 10000; i++ {
	//	body := randSeq(10000)
	//	title := randSeq(100)
	//
	//	postService.AddPost(context.Background(), int64(i%3), title, body, true)
	//
	//	time.Sleep(time.Millisecond * 10)
	//}

	wg := sync.WaitGroup{}
	for i := 0; i < 10000; i = i + 10 {
		go func() {
			wg.Add(1)
			for k := 0; k < 100; i++ {
				body := randSeq(1500)

				_, err := commentService.NewComment(context.Background(), int64(k%3+1), int64(i), body, nil)
				if err != nil {
					println("err: " + err.Error())
					println("user - ", int64(k%3))
				}

				time.Sleep(time.Millisecond * 10)
			}
			wg.Done()
		}()
	}

	wg.Wait()

}
