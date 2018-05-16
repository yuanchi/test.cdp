package test_cdp

import (
	"context"
	"time"
	"testing"
	//"log"
	"fmt"
	"os/exec"
	"strings"
	"errors"
	"io/ioutil"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/dom"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/rpcc"
)

func TestCDPSample(t *testing.T){
	cmd := exec.Command("/opt/google/chrome/google-chrome", "--headless", "--disable-gpu", "--no-sandbox", "--remote-debugging-port=9223")
	if err := cmd.Start(); err != nil {
		fmt.Printf("%v", err)
		return
	}
	defer cmd.Process.Kill()

	ctx, cancel := context.WithTimeout(context.Background(), 60 * time.Second)
	defer cancel()	

	devt := devtool.New("http://127.0.0.1:9223")
	pt, err := devt.Get(ctx, devtool.Page)
	if err != nil {
		ticker :=time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()
		s := "connection refused"

		timeout := time.After(5 * time.Second)

		loop:
		for{
			select{
				case <- ticker.C:
					fmt.Printf("retry connecting after 80 millisec.")
					pt ,err = devt.Create(ctx)
					if err != nil {
						if !strings.Contains(err.Error(),s){
							break loop	
						}
					}else{
						ticker.Stop()
						break loop
					}
				case <- timeout:
					err = errors.New("5 sec. timeout")	
					break loop
			}
		}
	}
	
	if err != nil {
		fmt.Printf("%v", err)
		return
	}
	fmt.Println("connected")

	// the above is the sample at the official github website
	conn, err := rpcc.DialContext(ctx, pt.WebSocketDebuggerURL)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}
	defer conn.Close()

	c := cdp.NewClient(conn)

	domContent, err := c.Page.DOMContentEventFired(ctx)
	if err != nil {
		fmt.Printf("%v", err)
		return	
	}
	defer domContent.Close()

	if err = c.Page.Enable(ctx); err != nil {
		fmt.Printf("%v", err)
		return
	}

	navArgs := page.NewNavigateArgs("https://www.google.com").
		SetReferrer("https://duckduckgo.com")
	nav, err := c.Page.Navigate(ctx, navArgs)
	if err != nil {
		fmt.Printf("%v", err)
		return	
	}

	if _, err = domContent.Recv(); err != nil {
		fmt.Printf("%v", err)
		return
	}

	fmt.Printf("Page loaded with frame ID: %s\n", nav.FrameID)

	doc, err := c.DOM.GetDocument(ctx, nil)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	result, err := c.DOM.GetOuterHTML(ctx, &dom.GetOuterHTMLArgs{
		NodeID: &doc.Root.NodeID,
	})
	if err != nil {
		fmt.Printf("%v", err)
		return
	}
	// if printing the outer html to the console, it may cause timeout
	fmt.Printf("HTML len: %d\n", len(result.OuterHTML))

	screenshotName := "screenshot.jpg"
	screenshotArgs := page.NewCaptureScreenshotArgs().
		SetFormat("jpeg").
		SetQuality(80)
	screenshot, err := c.Page.CaptureScreenshot(ctx, screenshotArgs)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}
	if err = ioutil.WriteFile(screenshotName, screenshot.Data, 0644); err != nil {
		fmt.Printf("%v", err)
		return
	}

	fmt.Printf("Saved screenshot: %s\n", screenshotName)
		
}
