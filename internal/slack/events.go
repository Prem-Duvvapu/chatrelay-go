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
    "go.opentelemetry.io/otel/trace"
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
                        tracer := otel.Tracer("chatrelay-tracer")
                        ctx, span := tracer.Start(ctx, "slack_listener",
                            trace.WithAttributes(
                                attribute.String("user_id", ev.User),
                                attribute.String("channel_id", ev.Channel),
                                attribute.String("query", ev.Text),
                            ),
                        )
                        defer span.End()

                        log.Printf("[Slack] Bot was mentioned by user %s in channel %s", ev.User, ev.Channel)

                        _, _, err := api.PostMessage(ev.Channel, slack.MsgOptionText("Processing your request...", false))
                        if err != nil {
                            log.Printf("Error sending message: %v\n", err)
                        }

                        go forwardToBackend(ctx, ev.User, ev.Text, ev.Channel, api)
                    }
                }

            default:
                log.Printf("Ignored event: %+v\n", evt.Type)
            }
        }
    }()

    client.Run()
}

func forwardToBackend(ctx context.Context, userID, query string, channelID string, api *slack.Client) {
    tracer := otel.Tracer("chatrelay-tracer")
    ctx, span := tracer.Start(ctx, "forwardToBackend")
    defer span.End()

    span.SetAttributes(
        attribute.String("user_id", userID),
        attribute.String("channel_id", channelID),
        attribute.String("query", query),
    )

    spanCtx := trace.SpanContextFromContext(ctx)
    log.Printf("[Backend] Forwarding to backend. trace_id=%s span_id=%s user_id=%s", spanCtx.TraceID(), spanCtx.SpanID(), userID)

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
    log.Printf("[Backend] Sent to backend. Status: %s", resp.Status)

    buf := make([]byte, 1024)
    for {
        n, err := resp.Body.Read(buf)
        if n > 0 {
            message := string(buf[:n])
            log.Printf("[Backend] Received data from backend: %s", message)

            _, _, err := api.PostMessage(channelID, slack.MsgOptionText(message, false))
            if err != nil {
                log.Printf("Error sending message to Slack: %v\n", err)
            }
        }

        if err != nil {
            break
        }
    }
}
