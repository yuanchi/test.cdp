package test_cdp

import (
	"os/exec"
	"os"
	"time"
	"io"
	"log"
	"fmt"
	"testing"
)

func TestStartChromeAndTimeoutExit(t *testing.T) {
	pr, pw := io.Pipe()
	defer pw.Close()

	cmd := exec.Command("/opt/google/chrome/google-chrome", "--headless", "--disable-gpu", "--no-sandbox", "--remote-debugging-port=9221", "--remote-debugging-address=192.168.1.43")
	cmd.Stdout = pw
	
	go func(){ // no output when startup google chrome
		defer pr.Close()
		if _, err := io.Copy(os.Stdout, pr); err != nil {
			log.Fatalln(err)
		}
		fmt.Println("pipe...")
	}()
	
	if err := cmd.Start(); err != nil {
		log.Fatalln(err)
	}

	timeout := time.After(10 * time.Second)
	select{
		case <- timeout:
			if err := cmd.Process.Kill(); err != nil {
				log.Fatalln(err)
			}
			fmt.Println("10 sec. timeout")
			return
	}		

}
