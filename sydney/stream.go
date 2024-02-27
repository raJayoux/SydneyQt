package sydney

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sydneyqt/util"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/tidwall/gjson"
	"nhooyr.io/websocket"
)

func (o *Sydney) AskStream(options AskStreamOptions) (<-chan Message, error) {
	out := make(chan Message)
	options.messageID = uuid.New().String()
	conversation, ch, err := o.AskStreamRaw(options)
	if err != nil {
		return nil, err
	}
	go func(out chan Message, ch <-chan RawMessage) {
		defer func() {
			slog.Info("AskStream is closing out message channel")
			close(out)
		}()
		wrote := 0
		sendSuggestedResponses := func(message gjson.Result) {
			if message.Get("suggestedResponses").Exists() {
				arr := util.Map(message.Get("suggestedResponses").Array(), func(v gjson.Result) string {
					return v.Get("text").String()
				})
				v, _ := json.Marshal(arr)
				out <- Message{
					Type: MessageTypeSuggestedResponses,
					Text: string(v),
				}
			}
		}
		var sourceAttributes []SourceAttribute
		for msg := range ch {
			if msg.Error != nil {
				slog.Error("Ask stream message", "error", msg.Error)
				if strings.Contains(msg.Error.Error(), "CAPTCHA") {
					if options.disableCaptchaBypass {
						err0 := errors.New("infinite CAPTCHA detected; " +
							"please resolve it manually on Bing's website or mobile client")
						out <- Message{
							Type:  MessageTypeError,
							Text:  err0.Error(),
							Error: err0,
						}
						return
					}
					slog.Info("Start to resolve the captcha", "server", o.bypassServer)
					out <- Message{
						Type: MessageTypeResolvingCaptcha,
						Text: "Please wait patiently while we are resolving the CAPTCHA...",
					}
					if o.bypassServer == "" {
						err = o.ResolveCaptcha(options.StopCtx)
					} else {
						err = o.BypassCaptcha(options.StopCtx, conversation.ConversationId,
							options.messageID)
					}
					if err != nil {
						if !errors.Is(err, context.Canceled) {
							err = fmt.Errorf("cannot resolve CAPTCHA automatically; "+
								"please resolve it manually on Bing's website or mobile client: %w", err)
							out <- Message{
								Type:  MessageTypeError,
								Text:  err.Error(),
								Error: err,
							}
						}
						return
					}
					newOptions := options
					newOptions.disableCaptchaBypass = true
					newOptions.messageID = ""
					newCh, err := o.AskStream(newOptions)
					if err != nil {
						out <- Message{
							Type:  MessageTypeError,
							Text:  err.Error(),
							Error: err,
						}
						return
					}
					for newMsg := range newCh { // proxy messages from recursive AskStream
						out <- newMsg
					}
					return
				} else {
					out <- Message{
						Type:  MessageTypeError,
						Text:  msg.Error.Error(),
						Error: msg.Error,
					}
					return
				}
			}
			data := gjson.Parse(msg.Data)
			if data.Get("type").Int() == 1 && data.Get("arguments.0.messages").Exists() {
				message := data.Get("arguments.0.messages.0")
				msgType := message.Get("messageType")
				messageText := message.Get("text").String()
				messageHiddenText := message.Get("hiddenText").String()
				contentOrigin := message.Get("contentOrigin").String()
				switch msgType.String() {
				case "InternalSearchQuery":
					out <- Message{
						Type: MessageTypeSearchQuery,
						Text: messageText,
					}
				case "InternalSearchResult":
					if strings.Contains(messageHiddenText,
						"Web search returned no relevant result") {
						slog.Info("Web search returned no relevant result")
						continue
					}
					if !gjson.Valid(messageText) {
						slog.Error("Error when parsing InternalSearchResult", "messageText", messageText)
						continue
					}
					arr := gjson.Parse(messageText).Array()
					for _, group := range arr {
						group.ForEach(func(key, value gjson.Result) bool {
							for _, subGroup := range value.Array() {
								sourceAttributes = append(sourceAttributes, SourceAttribute{
									Link:  subGroup.Get("url").String(),
									Title: subGroup.Get("title").String(),
								})
							}
							return true
						})
					}
				case "InternalLoaderMessage":
					if message.Get("hiddenText").Exists() {
						out <- Message{
							Type: MessageTypeLoading,
							Text: messageHiddenText,
						}
						continue
					}
					if message.Get("text").Exists() {
						out <- Message{
							Type: MessageTypeLoading,
							Text: messageText,
						}
						continue
					}
					out <- Message{
						Type: MessageTypeLoading,
						Text: message.Raw,
					}
				case "GenerateContentQuery":
					if message.Get("contentType").String() != "IMAGE" {
						continue
					}
					generativeImage := GenerativeImage{
						Text: messageText,
						URL: "https://www.bing.com/images/create?" +
							"partner=sydney&re=1&showselective=1&sude=1&kseed=7500&SFX=2&gptexp=unknown" +
							"&q=" + url.QueryEscape(messageText) + "&iframeid=" +
							message.Get("messageId").String(),
					}
					v, err := json.Marshal(&generativeImage)
					if err != nil {
						util.GracefulPanic(err)
					}
					out <- Message{
						Type: MessageTypeGenerativeImage,
						Text: string(v),
					}
				case "Progress":
					switch contentOrigin {
					case "CodeInterpreter":
						invocation := message.Get("invocation").String()
						if invocation == "" {
							continue
						}
						out <- Message{
							Type: MessageTypeExecutingTask,
							Text: invocation,
						}
					default:
						slog.Warn("Unsupported progress type",
							"contentOrigin", contentOrigin,
							"triggered-by", options.Prompt, "response", message.Raw)
					}
				case "GeneratedCode":
					out <- Message{
						Type: MessageTypeGeneratedCode,
						Text: messageText,
					}
				case "":
					if data.Get("arguments.0.cursor").Exists() {
						wrote = 0
						// extract search result from text block
						if text := strings.TrimSuffix(message.Get("adaptiveCards.0.body.0.text").String(),
							messageText); strings.TrimSpace(text) != "" {
							arr := lo.Filter(lo.Map(strings.Split(text, "\n"), func(item string, index int) string {
								return strings.Trim(item, " \"")
							}), func(item string, index int) bool {
								return item != ""
							})
							re := regexp.MustCompile(`\[(\d+)]: (.*)`)
							var resultSources []SourceAttribute
							for _, line := range arr {
								matches := re.FindStringSubmatch(line)
								if len(matches) == 0 {
									continue
								}
								ix := matches[1]
								link := matches[2]
								sourceAttribute, ok := lo.Find(sourceAttributes, func(item SourceAttribute) bool {
									return item.Link == link
								})
								if !ok {
									continue
								}
								sourceAttribute.Index, _ = strconv.Atoi(ix)
								resultSources = append(resultSources, sourceAttribute)
							}
							var resultArr []string
							for _, src := range resultSources {
								v, _ := json.Marshal(&src)
								resultArr = append(resultArr, "  "+string(v))
							}
							if len(resultArr) != 0 {
								out <- Message{
									Type: MessageTypeSearchResult,
									Text: "[\n" + strings.Join(resultArr, ",\n") + "\n]",
								}
							}
						}
					}
					if contentOrigin == "Apology" {
						if wrote != 0 {
							out <- Message{
								Type:  MessageTypeError,
								Text:  "Message revoke detected",
								Error: ErrMessageRevoke,
							}
						} else {
							out <- Message{
								Type:  MessageTypeError,
								Text:  "Looks like the user's message has triggered the Bing filter",
								Error: ErrMessageFiltered,
							}
						}
						return
					} else {
						if wrote < len(messageText) {
							out <- Message{
								Type: MessageTypeMessageText,
								Text: messageText[wrote:],
							}
							wrote = len(messageText)
						} else if wrote > len(messageText) { // Bing deletes some already sent text
							wrote = len(messageText)
						}
						sendSuggestedResponses(message)
					}
				default:
					slog.Warn("Unsupported message type",
						"type", msgType.String(), "triggered-by", options.Prompt, "response", message.Raw)
				}
			} else if data.Get("type").Int() == 2 && data.Get("item.messages").Exists() {
				message := data.Get("item.messages|@reverse|0")
				sendSuggestedResponses(message)
			}
		}
	}(out, ch)
	return out, nil
}
func (o *Sydney) AskStreamRaw(options AskStreamOptions) (CreateConversationResponse, <-chan RawMessage, error) {
	slog.Info("AskStreamRaw called, creating conversation...")
	conversation, err := o.createConversation()
	if err != nil {
		return CreateConversationResponse{}, nil, err
	}
	select {
	case <-options.StopCtx.Done():
		return conversation, nil, options.StopCtx.Err()
	default:
	}
	msgChan := make(chan RawMessage)
	go func(msgChan chan RawMessage) {
		defer func(msgChan chan RawMessage) {
			slog.Info("AskStreamRaw is closing raw message channel")
			close(msgChan)
		}(msgChan)
		client, _, err := util.MakeHTTPClient(o.proxy, 0)
		if err != nil {
			msgChan <- RawMessage{
				Error: err,
			}
			return
		}
		messageID := options.messageID
		if messageID == "" {
			msgID, err := uuid.NewUUID()
			if err != nil {
				msgChan <- RawMessage{
					Error: err,
				}
				return
			}
			messageID = msgID.String()
		}
		httpHeaders := http.Header{}
		for k, v := range o.headers() {
			httpHeaders.Set(k, v)
		}
		ctx, cancel := util.CreateTimeoutContext(10 * time.Second)
		defer cancel()
		connRaw, resp, err := websocket.Dial(ctx,
			o.wssURL+util.Ternary(conversation.SecAccessToken != "", "?sec_access_token="+
				url.QueryEscape(conversation.SecAccessToken), ""),
			&websocket.DialOptions{
				HTTPClient: client,
				HTTPHeader: httpHeaders,
			})
		if err != nil {
			msgChan <- RawMessage{
				Error: err,
			}
			return
		}
		if resp.StatusCode != 101 {
			msgChan <- RawMessage{
				Error: errors.New("cannot establish a websocket connection"),
			}
			return
		}
		defer connRaw.CloseNow()
		select {
		case <-options.StopCtx.Done():
			slog.Info("Exit askStream because of received signal from stopCtx")
			return
		default:
		}
		connRaw.SetReadLimit(-1)
		conn := &Conn{Conn: connRaw, debug: o.debug}
		err = conn.WriteWithTimeout([]byte(`{"protocol": "json", "version": 1}`))
		if err != nil {
			msgChan <- RawMessage{
				Error: err,
			}
			return
		}
		conn.ReadWithTimeout()
		err = conn.WriteWithTimeout([]byte(`{"type": 6}`))
		if err != nil {
			msgChan <- RawMessage{
				Error: err,
			}
			return
		}
		chatMessage := ChatMessage{
			Arguments: []Argument{
				{
					OptionsSets:         o.optionsSet,
					Source:              "cib-ccp",
					AllowedMessageTypes: o.allowedMessageTypes,
					SliceIds:            o.sliceIDs,
					Verbosity:           "verbose",
					Scenario:            "SERP",
					TraceId:             util.MustGenerateRandomHex(16),
					RequestId:           messageID,
					IsStartOfSession:    true,
					Message: ArgumentMessage{
						Locale: o.locale,
						Market: o.locale,
						Region: "US",
						Location: fmt.Sprintf("lat:%.6f;long:%.6f;re=1000m;",
							o.locationHint.Center.Latitude,
							o.locationHint.Center.Longitude),
						LocationHints: []LocationHint{
							o.locationHint,
						},
						Author:      "user",
						InputMethod: "Keyboard",
						Text:        options.Prompt,
						MessageType: []string{"Chat", "CurrentWebpageContextRequest"}[util.RandIntInclusive(0, 1)],
						RequestId:   messageID,
						MessageId:   messageID,
						ImageUrl:    util.Ternary[any](options.ImageURL == "", nil, options.ImageURL),
					},
					Tone: o.conversationStyle,
					ConversationSignature: util.Ternary[any](conversation.ConversationSignature == "",
						nil, conversation.ConversationSignature),
					Participant:    Participant{Id: conversation.ClientId},
					SpokenTextMode: "None",
					ConversationId: conversation.ConversationId,
					PreviousMessages: []PreviousMessage{
						{
							Author:      "user",
							Description: options.WebpageContext,
							ContextType: "WebPage",
							MessageType: "Context",
						},
					},
					GptId: o.gptID,
				},
			},
			InvocationId: "0",
			Target:       "chat",
			Type:         4,
		}
		chatMessageV, err := json.Marshal(&chatMessage)
		if err != nil {
			msgChan <- RawMessage{
				Error: err,
			}
			return
		}
		err = conn.WriteWithTimeout(chatMessageV)
		if err != nil {
			msgChan <- RawMessage{
				Error: err,
			}
			return
		}
		for {
			select {
			case <-options.StopCtx.Done():
				slog.Info("Exit askStream because of received signal from stopCtx")
				return
			default:
			}
			messages, err := conn.ReadWithTimeout()
			if err != nil {
				msgChan <- RawMessage{
					Error: err,
				}
				return
			}
			if time.Now().Unix()%6 == 0 {
				err = conn.WriteWithTimeout([]byte(`{"type": 6}`))
				if err != nil {
					msgChan <- RawMessage{
						Error: err,
					}
					return
				}
			}
			for _, msg := range messages {
				if msg == "" {
					continue
				}
				if !gjson.Valid(msg) {
					msgChan <- RawMessage{
						Error: errors.New("malformed json"),
					}
					return
				}
				result := gjson.Parse(msg)
				if result.Get("type").Int() == 2 && result.Get("item.result.value").String() != "Success" {
					msgChan <- RawMessage{
						Error: errors.New("bing explicit error: value: " +
							result.Get("item.result.value").String() + "; message: " +
							result.Get("item.result.message").String()),
					}
					return
				}
				msgChan <- RawMessage{
					Data: msg,
				}
				if result.Get("type").Int() == 2 {
					// finish the conversation
					return
				}
			}
		}
	}(msgChan)
	return conversation, msgChan, nil
}
