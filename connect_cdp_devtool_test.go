package test_cdp

import (
	"context"
	"time"
	"testing"
	"log"
	"fmt"
	"os/exec"
	"strings"
	"errors"

	//"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	//"github.com/mafredri/cdp/protocol/dom"
	//"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/rpcc"
)

func TestConnectCDPDevTool(t *testing.T){
	cmd := exec.Command("/opt/google/chrome/google-chrome", "--headless", "--disable-gpu", "--no-sandbox", "--remote-debugging-port=9222", "--remote-debugging-address=127.0.0.1")
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	defer cmd.Process.Kill()

	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()	

	devt := devtool.New("http://127.0.0.1:9222")
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
					fmt.Println("retry connecting after 80 millisec.")
					pt ,err = devt.Create(ctx)
					if err != nil {
						if !strings.Contains(err.Error(),s){
							break loop	
						}
					}else{
						err = nil
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
		log.Fatal(err)
	}
	fmt.Println("connected")
	conn, err := rpcc.DialContext(ctx, pt.WebSocketDebuggerURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
}
