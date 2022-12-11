package main

import (
	"context"
	"os"
	"path"

	"log"

	"github.com/chromedp/chromedp"
)

type archiveTask struct {
	targetURL      string
	targetBaseName string
}

func createArciveWorker(taskChan chan archiveTask) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var buf []byte
	for task := range taskChan {
		err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.EmulateViewport(1920, 1080),
			chromedp.Navigate(task.targetURL),
			chromedp.FullScreenshot(&buf, 100),
		})
		if err != nil {
			log.Printf("failed to take screenshot: %v", err)
		} else {
			if err := os.WriteFile(path.Join(*screenshotFolder, task.targetBaseName+".png"), buf, 0644); err != nil {
				log.Printf("failed to store screenshot: %v", err)
			}
		}
	}
}
