package slack

import (
    "bytes"
    "context"
    "encoding/json"
    "log"
    "net/http"
    "os"

    "github.com/slack-go/slack"
    "github.com/slack-go/slack/slackevents"
    "github.com/slack-go/slack/socketmode"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
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

                        query := ev.Text
                        userID := ev.User
                        channelID := ev.Channel

                        _, _, err := api.PostMessage(channelID, slack.MsgOptionText("Processing your request...", false))
                        if err != nil {
                            log.Printf("Error sending message: %v\n", err)
                        }

                        go forwardToBackend(userID, query, channelID, api)
                    }
                }

            default:
                log.Printf("Ignored event: %+v\n", evt.Type)
            }
        }
    }()

    client.Run()
}

func forwardToBackend(userID, query string, channelID string, api *slack.Client) {
    ctx := context.Background()
    tracer := otel.Tracer("chatrelay-tracer")
    ctx, span := tracer.Start(ctx, "forwardToBackend") // ⬅️ start a span
    defer span.End()

    backendURL := os.Getenv("BACKEND_URL")
    if backendURL == "" {
        log.Println("BACKEND_URL is not set")
        return
    }

    payload := map[string]string{
        "user_id": userID,
        "query":   query,
    }

    jsonData, err := json.Marshal(payload)
    if err != nil {
        span.RecordError(err) // ⬅️ record error in trace
        log.Printf("Error marshaling request: %v\n", err)
        return
    }

    resp, err := http.Post(backendURL, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        log.Printf("Error sending request to backend: %v\n", err)
        return
    }
    defer resp.Body.Close()

    span.SetAttributes(attribute.String("http.status", resp.Status))

    log.Printf("Sent to backend. Status: %s\n", resp.Status)

    // Now read the SSE stream and send messages to Slack
    buf := make([]byte, 1024) // Adjust the buffer size if needed
    for {
        n, err := resp.Body.Read(buf)
        if err != nil && err.Error() != "EOF" {
            log.Printf("Error reading from SSE stream: %v\n", err)
            break
        }

        if n > 0 {
            message := string(buf[:n])
            log.Printf("Received data from backend: %s", message)

            // Send the message to Slack
            _, _, err := api.PostMessage(channelID, slack.MsgOptionText(message, false))
            if err != nil {
                log.Printf("Error sending message to Slack: %v\n", err)
            }
        }

        if err != nil && err.Error() == "EOF" {
            break
        }
    }
}
