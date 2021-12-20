package wk3

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
)

const (
	LocalHttp = "127.0.0.1:8080"
)

func HelloCamp(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hollo go Camp")
}

func main() {
	ctx := context.Background()
	ctx, cancelFunc := context.WithCancel(ctx)
	g, errCtx := errgroup.WithContext(ctx)

	srv := &http.Server{
		Addr: LocalHttp,
	}
	http.HandleFunc("/hello", HelloCamp)

	g.Go(func() error {
		fmt.Printf("Http Server %s starts", LocalHttp)
		if err := srv.ListenAndServe(); err != nil {
			fmt.Printf("there is an error from server %s, error: %s", LocalHttp, err)
			return err
		}
		select {
		case <-errCtx.Done():
			fmt.Printf("Need to stop server %s", LocalHttp)
			srv.Shutdown(errCtx)
		}
		return nil
	})

	ch := make(chan os.Signal, 1)
	signal.Notify(ch)
	g.Go(func() error {
		fmt.Printf("Waiting for OS signal")
		select {
		case <-errCtx.Done():
			return errCtx.Err()
		case <-ch:
			cancelFunc()
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		fmt.Printf("group wait error '%s'", err.Error())
	}
	close(ch)
	return
}
