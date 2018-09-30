package main

import (
	"fmt"
	"os/exec"

	"github.com/rasky/realcrypto/common"
)

func main() {
	go common.PollTimezone()

	fmt.Println("Premi ENTER per suonare l'audio...")
	fmt.Scanln()

	t := common.NowAtIpLoc()
	if t.Hour() >= 9 && t.Hour() <= 20 {
		fmt.Printf("Qui sono le %02d:%02d, quindi posso fare casino\n", t.Hour(), t.Minute())
		exec.Command("play", "modem.ogg").Run()
	} else {
		fmt.Printf("Qui sono le %02d:%02d, meglio far silenzio\n", t.Hour(), t.Minute())
	}
}
