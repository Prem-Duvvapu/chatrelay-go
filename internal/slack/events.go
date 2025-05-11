package slack

import (
    "context"
    "log"
    "os"

    "github.com/slack-go/slack"
    "github.com/slack-go/slack/slackevents"
    "github.com/slack-go/slack/socketmode"
)

func StartSlackListener(ctx context.Context) {
    appToken := os.Getenv("SLACK_APP_TOKEN")
    botToken := os.Getenv("SLACK_BOT_TOKEN")

    api := slack.New(
        botToken,
        slack.OptionDebug(true),
        slack.OptionAppLevelToken(appToken),
    )

    client := socketmode.New(api, socketmode.OptionDebug(true))

    go func() {
        for evt := range client.Events {
            switch evt.Type {
            case socketmode.EventTypeEventsAPI:
                eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
                if !ok {
                    log.Printf("Ignored %+v\n", evt)
                    continue
                }

                client.Ack(*evt.Request)

                if eventsAPIEvent.Type == slackevents.CallbackEvent {
                    innerEvent := eventsAPIEvent.InnerEvent
                    switch ev := innerEvent.Data.(type) {
                    case *slackevents.AppMentionEvent:
                        log.Println("Bot was mentioned!")
                        _, _, err := api.PostMessage(ev.Channel, slack.MsgOptionText("Hello, I'm ChatRelay! 👋", false))
                        if err != nil {
                            log.Printf("Error sending message: %v\n", err)
                        }
                    }
                }

            default:
                log.Printf("Ignored event: %+v\n", evt.Type)
            }
        }
    }()

    client.Run()
}
