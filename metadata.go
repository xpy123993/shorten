package main

import (
	"context"
	"os"
	"path"

	"log"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type archiveTask struct {
	targetURL      string
	targetBaseName string
}

func createArciveWorker(taskChan chan archiveTask) {
	for task := range taskChan {
		ctx, cancel := chromedp.NewContext(context.Background())
		var screenBuf []byte
		var pdfBuf []byte
		err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.EmulateViewport(1920, 1080),
			chromedp.Navigate(task.targetURL),
			chromedp.FullScreenshot(&screenBuf, 100),
			chromedp.ActionFunc(func(ctx context.Context) error {
				buf, _, err := page.PrintToPDF().WithPrintBackground(false).Do(ctx)
				if err != nil {
					return err
				}
				pdfBuf = buf
				return nil
			}),
		})
		if err != nil {
			log.Printf("failed to take screenshot: %v", err)
		} else {
			if err := os.WriteFile(path.Join(*metadataFolder, task.targetBaseName+".png"), screenBuf, 0644); err != nil {
				log.Printf("failed to store screenshot: %v", err)
			}
			if err := os.WriteFile(path.Join(*metadataFolder, task.targetBaseName+".pdf"), pdfBuf, 0644); err != nil {
				log.Printf("failed to store PDF: %v", err)
			}
		}
		cancel()
	}
}
