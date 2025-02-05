package chat

import (
	"bufio"
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
	"golang.org/x/term"

	"micheam.com/aico/internal/config"
	"micheam.com/aico/internal/openai"
	"micheam.com/aico/internal/openai/chat"
	"micheam.com/aico/internal/openai/models"
	"micheam.com/aico/internal/spinner"
	"micheam.com/aico/internal/theme"
)

var (
	// Spinner settings
	spinnerFrames   = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinnerInterval = 100 * time.Millisecond
	authEnvKey      = "OPENAI_API_KEY" // TODO: move to config
)

type Handler struct {
	cfg     *config.Config
	persona *config.Personality

	logger *slog.Logger
}

func New(cfg *config.Config, logger *slog.Logger) *Handler {
	return &Handler{cfg: cfg, logger: logger}
}

// Run starts the chat.
func (h *Handler) Run(ctx context.Context) error {
	// Prepare message with persona
	messages := make([]openai.Message, 0)

	// System messages from persona
	for _, msg := range h.persona.Messages {
		messages = append(messages, &openai.SystemMessage{Content: msg})
	}

	// Spinner settings
	spinner := spinner.New(spinnerInterval, spinnerFrames)

	prompt := "> "

	authToken := os.Getenv(authEnvKey)
	if authToken == "" {
		fmt.Printf(theme.Error("%s is not set"), authEnvKey)
		os.Exit(1)
	}

	client := chat.New(authToken)
	model := models.Model(h.cfg.Chat.Model)
	if !chat.IsAvailableModel(model) {
		return fmt.Errorf("model %s is not available", model)
	}

	fmt.Printf(theme.Info("Conversation with %s\n"), model)
	fmt.Println(theme.Info("------------------------------"))
	logger := h.logger.With("model", model.String())

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(theme.Info(prompt))
		text, _ := reader.ReadString('\n')
		text = strings.ReplaceAll(text, "\n", "") // convert CRLF to LF

		switch text {

		default: // store user input
			logger.Debug("User input", "text", text)
			if strings.HasPrefix(text, "SYSTEM:") {
				messages = append(messages, &openai.SystemMessage{Content: text})
			} else {
				// Handle image URL
				if strings.HasPrefix(text, "<") && strings.HasSuffix(text, ">") {
					urlStr := text[1 : len(text)-1]
					if u, err := url.Parse(urlStr); err == nil {
						messages = append(messages, &openai.UserMessage{
							Content: []openai.Content{&openai.ImageContent{URL: *u}},
						})
					}
					continue
				}
				// Handle plain text
				messages = append(messages, &openai.UserMessage{
					Content: []openai.Content{&openai.TextContent{Text: text}},
				})
			}

		case "": // empty input
			continue

		case ".quit", ".q", ".exit":
			return nil

		case ".send", ";;":
			fmt.Println()
			spinner.Start()
			defer spinner.Stop()

			req := chat.NewRequest(messages, chat.WithModel(model))
			logger.Debug("ChatRequest", "req", req)

			// width of terminal
			width, _, err := term.GetSize(0)
			if err != nil {
				width = 100
			}

			var cnt int // Current width of output
			content := new(strings.Builder)

			if err := client.DoStream(ctx, req, func(resp *chat.Response) error { // Block until completion DONE
				spinner.Stop()
				delta := resp.Choices[0].Delta
				if delta == nil {
					return nil
				}

				deltaContent := delta.Content
				deltaContent = strings.ReplaceAll(deltaContent, "\t", "  ") // convert tab to 2 spaces
				wrapped := strings.Contains(deltaContent, "\n")
				deltaWidth := runewidth.StringWidth(strings.ReplaceAll(deltaContent, "\n", ""))

				if wrapped {
					cnt = 0
				}
				if cnt+deltaWidth > width {
					fmt.Printf("\n")
					cnt = 0
				}

				fmt.Printf(theme.Reply("%s"), delta.Content)
				cnt += runewidth.StringWidth(delta.Content)

				_, err := content.WriteString(delta.Content)

				return err
			}); err != nil {
				fmt.Printf(theme.Error("ChatCompletion error: %v\n"), err)
				logger.Debug("ChatCompletion error", "err", err)
				spinner.Stop()
				continue
			}
			fmt.Printf("\n\n")

			messages = append(messages, &openai.AssistantMessage{
				Content: []openai.Content{&openai.TextContent{Text: content.String()}},
			})
		}
	}
}

func (h *Handler) WithPersona(p *config.Personality) *Handler {
	h.persona = p
	return h
}

func (h *Handler) Persona() *config.Personality {
	if h.persona == nil {
		return h.cfg.Chat.GetDefaultPersona()
	}
	return h.persona
}
