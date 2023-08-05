package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pills-of-cs/bot"
)

func main() {

	var err error
	ctx := context.Background()

	bot, client, err := bot.NewBotWithConfig()
	if err != nil {
		log.Fatalf("got error: %v", err)
		os.Exit(1)
	}

	// Create the db and the collection
	bot.UserRepo.Client = client

	err = bot.Bot.Run()
	if err != nil {
		log.Fatalf("got error: %v", err)
		os.Exit(1)
	}
	go bot.Start(ctx)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}

/*
2023/08/05 19:23:28 | Update				message |Parsed
2023/08/05 19:23:28 | Update				message |Parsed
panic: runtime error: index out of range [1] with length 1
goroutine 45 [running]:
github.com/pills-of-cs/bot.Bot.SchedulePill({0xc0005ba7b8?}, {0xcc3570?, 0xc000188000}, 0xc0001ba100)
/app/bot/handlers.go:152 +0x4fd
github.com/pills-of-cs/bot.(*Bot).Start.func7(0x0?)
/app/bot/bot.go:209 +0x3b
created by github.com/SakoDroid/telego/parser.checkTextMsgHandlers
/go/pkg/mod/github.com/!sako!droid/telego@v1.8.0/parser/handlers.go:61 +0xd8
2023/08/05 19:23:30
&{{0xc00020a880 false 0x4d3b00 0xc000206330 0xc000206348} 0xc000200270 0xc000204cc0}
[Info]: Database connection established
2023/08/05 19:23:30 | getWebhookInfo			Success |570124µs
2023/08/05 19:23:31 | Update				message |Parsed
2023/08/05 19:23:31 [time.Parse]: parsed time: 0000-01-01 20:36:00 +0000 UTC

*/
