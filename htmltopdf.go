package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"sync"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// starts a file server of a certain dir on a certain port
// returns a close function
func startFileServing(dir string, port int) func() {
	srv := &http.Server{Addr: fmt.Sprintf(":%d", port)}

	http.Handle("/", http.FileServer(http.Dir(dir)))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done() // let main know we are done cleaning up

		// always returns error. ErrServerClosed on graceful close
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// unexpected error. port in use?
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	return func() {
		srv.Shutdown(context.TODO())
		if err := srv.Shutdown(context.TODO()); err != nil {
			panic(err) // failure/timeout shutting down the server gracefully
		}

		// wait for goroutine started in startHttpServer() to stop
		// NOTE: as @sander points out in comments, this might be unnecessary.
		wg.Wait()
		// returning reference so caller can call Shutdown()
	}
}

type Page struct {
	PageNum int
	PDF     []byte
}

func savePagesToPdf(url, outputFolder string) error {
	err := os.MkdirAll(outputFolder, 0o755)
	if err != nil {
		return err
	}
	opts := []chromedp.ExecAllocatorOption{
		chromedp.Headless, // chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.DisableGPU,
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("enable-features", "NetworkService,NetworkServiceInProcess"),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-breakpad", true),
		chromedp.Flag("disable-client-side-phishing-detection", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-features", "site-per-process,Translate,BlinkGenPropertyTrees"),
		chromedp.Flag("disable-hang-monitor", true),
		chromedp.Flag("disable-ipc-flooding-protection", true),
		chromedp.Flag("disable-popup-blocking", true),
		chromedp.Flag("disable-prompt-on-repost", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("force-color-profile", "srgb"),
		chromedp.Flag("metrics-recording-only", true),
		chromedp.Flag("safebrowsing-disable-auto-update", true),
		chromedp.Flag("enable-automation", true),
		chromedp.Flag("password-store", "basic"),
		chromedp.Flag("use-mock-keychain", true),
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// also set up a custom logger
	taskCtx, cancel := chromedp.NewContext(allocCtx) // chromedp.WithLogf(log.Printf), chromedp.WithDebugf(log.Printf), chromedp.WithErrorf(log.Printf)
	defer cancel()

	pages := make(chan Page)

	// capture pdf
	go func() {
		chromedp.Run(taskCtx, printToPDF(url, pages))
	}()

	for page := range pages {
		if err := os.WriteFile(path.Join(outputFolder, fmt.Sprintf("%d.pdf", page.PageNum)), page.PDF, 0o644); err != nil {
			log.Println(err)
		}
	}
	return nil
}

// print a specific pdf page.
func printToPDF(urlstr string, pages chan Page) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.ActionFunc(waitLoaded),
		chromedp.ActionFunc(func(ctx context.Context) error {
			pageNum := 1
			for {
				log.Printf("printing page %d of %s to pdf\n", pageNum, urlstr)
				buf, _, err := page.PrintToPDF().WithPrintBackground(false).WithPageRanges(fmt.Sprintf("%v", pageNum)).Do(ctx)
				if err != nil {
					log.Println("ERROR PRINTING TO PDF", err)
					break
				}
				p := Page{PageNum: pageNum, PDF: buf}
				pages <- p
				pageNum++
			}
			close(pages)
			return nil
		}),
	}
}

// waitLoaded blocks until a target receives a Page.loadEventFired.
// https://github.com/chromedp/chromedp/issues/431
func waitLoaded(ctx context.Context) error {
	// TODO: this function is inherently racy, as we don't run ListenTarget
	// until after the navigate action is fired. For example, adding
	// time.Sleep(time.Second) at the top of this body makes most tests hang
	// forever, as they miss the load event.
	//
	// However, setting up the listener before firing the navigate action is
	// also racy, as we might get a load event from a previous navigate.
	//
	// For now, the second race seems much more common in real scenarios, so
	// keep the first approach. Is there a better way to deal with this?
	chromedp.Run(ctx,
		chromedp.Tasks{
			enableLifeCycleEvents(),
		},
	)
	ch := make(chan struct{})
	cctx, cancel := context.WithCancel(ctx)
	chromedp.ListenTarget(cctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *page.EventLifecycleEvent:
			if e.Name == `networkIdle` {
				cancel()
				close(ch)
			}
		}
	})
	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func enableLifeCycleEvents() chromedp.ActionFunc {
	return func(ctx context.Context) error {
		err := page.Enable().Do(ctx)
		if err != nil {
			return err
		}
		err = page.SetLifecycleEventsEnabled(true).Do(ctx)
		if err != nil {
			return err
		}
		return nil
	}
}
