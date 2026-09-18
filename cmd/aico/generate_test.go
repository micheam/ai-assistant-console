package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
	"micheam.com/aico/internal/assistant"
)

func TestReadSource_WithContent(t *testing.T) {
	input := "hello world"
	r := strings.NewReader(input)

	result, err := readSource(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != input {
		t.Errorf("expected %q, got %q", input, result)
	}
}

func TestReadSource_EmptyInput(t *testing.T) {
	r := strings.NewReader("")

	result, err := readSource(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestParseLabel(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		wantLabel string
		wantSpec  string
	}{
		{"no colon", "plain text", "", "plain text"},
		{"bare @file", "@path/to/file.go", "", "@path/to/file.go"},
		{"bare @-", "@-", "", "@-"},
		{"labeled @file", "buffer.go:@/tmp/buffer.go", "buffer.go", "@/tmp/buffer.go"},
		{"labeled @-", "buffer.go:@-", "buffer.go", "@-"},
		{"windows drive, unlabeled", `@C:\path\to\file.go`, "", `@C:\path\to\file.go`},
		{"windows drive, labeled", `win:@C:\path\to\file.go`, "win", `@C:\path\to\file.go`},
		{"colon in inline text is not a label", "https://example.com", "", "https://example.com"},
		{"go-doc-like inline text with colon", "Note: see below", "", "Note: see below"},
		{"colon with no following @ is not a label", "12:30", "", "12:30"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLabel, gotSpec := parseLabel(tt.raw)
			if gotLabel != tt.wantLabel || gotSpec != tt.wantSpec {
				t.Errorf("parseLabel(%q) = (%q, %q), want (%q, %q)",
					tt.raw, gotLabel, gotSpec, tt.wantLabel, tt.wantSpec)
			}
		})
	}
}

func TestResolveSource_DirectString(t *testing.T) {
	input := "function foo() { return 42; }"

	stdinConsumed := false
	block, file, name, err := resolveSource(input, "", &stdinConsumed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "<source>\n" + input + "\n</source>"
	if block != expected {
		t.Errorf("expected %q, got %q", expected, block)
	}
	if file != "" || name != "" {
		t.Errorf("expected no file/name for inline source, got file=%q name=%q", file, name)
	}
}

func TestResolveSource_FromFile(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")
	content := "package main\n\nfunc main() {}\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	stdinConsumed := false
	block, file, name, err := resolveSource("@"+tmpFile, "", &stdinConsumed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(block, "<source file=") {
		t.Errorf("expected result to contain '<source file=', got %q", block)
	}
	if !strings.Contains(block, content) {
		t.Errorf("expected result to contain file content, got %q", block)
	}
	if !strings.Contains(block, "</source>") {
		t.Errorf("expected result to contain '</source>', got %q", block)
	}
	if file != tmpFile {
		t.Errorf("expected file %q, got %q", tmpFile, file)
	}
	if name != "" {
		t.Errorf("expected no name for unlabeled @file source, got %q", name)
	}
}

func TestResolveSource_FileNotFound(t *testing.T) {
	stdinConsumed := false
	_, _, _, err := resolveSource("@/nonexistent/file.go", "", &stdinConsumed)
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestResolveSource_Stdin(t *testing.T) {
	stdinConsumed := false
	block, file, name, err := resolveSource("@-", "piped content", &stdinConsumed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !stdinConsumed {
		t.Error("expected stdinConsumed to be true after resolving @-")
	}
	want := "<source>\npiped content\n</source>"
	if block != want {
		t.Errorf("expected %q, got %q", want, block)
	}
	if file != "" || name != "" {
		t.Errorf("expected no file/name for unlabeled @-, got file=%q name=%q", file, name)
	}
}

func TestResolveSource_LabeledStdin(t *testing.T) {
	stdinConsumed := false
	block, file, name, err := resolveSource("buffer.go:@-", "piped content", &stdinConsumed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(block, `name="buffer.go"`) {
		t.Errorf("expected result to contain name attribute, got %q", block)
	}
	if file != "" {
		t.Errorf("expected no file for labeled stdin source, got %q", file)
	}
	if name != "buffer.go" {
		t.Errorf("expected name %q, got %q", "buffer.go", name)
	}
}

func TestResolveSource_StdinAlreadyConsumed(t *testing.T) {
	stdinConsumed := true
	_, _, _, err := resolveSource("@-", "piped content", &stdinConsumed)
	if err == nil {
		t.Error("expected error when stdin (@-) is referenced a second time")
	}
}

func TestDetectSource_ImplicitStdinFallback(t *testing.T) {
	stdinConsumed := false
	block, file, name, err := detectSource("", "piped content", &stdinConsumed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "<source>\npiped content\n</source>"
	if block != want {
		t.Errorf("expected %q, got %q", want, block)
	}
	if file != "" || name != "" {
		t.Errorf("expected no file/name for implicit stdin source, got file=%q name=%q", file, name)
	}
	if !stdinConsumed {
		t.Error("expected stdinConsumed to be true after falling back to implicit stdin")
	}
}

func TestDetectSource_NoSourceNoStdin(t *testing.T) {
	stdinConsumed := false
	block, _, _, err := detectSource("", "", &stdinConsumed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if block != "" {
		t.Errorf("expected empty block, got %q", block)
	}
}

func TestDetectSource_FlagAndStdin_Errors(t *testing.T) {
	stdinConsumed := false
	_, _, _, err := detectSource("inline text", "piped content", &stdinConsumed)
	if err == nil {
		t.Error("expected error when --source and piped stdin are both present")
	}
}

func TestDetectSource_ExplicitAtDash_DoesNotErrorAgainstItself(t *testing.T) {
	stdinConsumed := false
	block, _, _, err := detectSource("@-", "piped content", &stdinConsumed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(block, "piped content") {
		t.Errorf("expected block to contain piped content, got %q", block)
	}
}

func TestDetectSource_ContextClaimedStdin_SourceFallsBackToNothing(t *testing.T) {
	// Simulates --context @- having already consumed stdin before --source
	// is resolved: the implicit stdin-to-source fallback must not fire.
	stdinConsumed := true
	block, _, _, err := detectSource("", "piped content", &stdinConsumed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if block != "" {
		t.Errorf("expected empty block once stdin is already consumed, got %q", block)
	}
}

func TestUserMessage_WithSource(t *testing.T) {
	source := "<source>\nhello world\n</source>"
	prompt := "summarize this"

	userContents := make([]assistant.MessageContent, 0, 2)
	userContents = append(userContents, assistant.NewTextContent(source))
	userContents = append(userContents, assistant.NewTextContent(prompt))
	userMsg := assistant.NewUserMessage(userContents...)

	contents := userMsg.GetContents()
	if len(contents) != 2 {
		t.Fatalf("expected 2 contents, got %d", len(contents))
	}
	src, ok := contents[0].(*assistant.TextContent)
	if !ok {
		t.Fatalf("expected first content to be *TextContent")
	}
	if src.Text != source {
		t.Errorf("expected source text %q, got %q", source, src.Text)
	}
	msg, ok := contents[1].(*assistant.TextContent)
	if !ok {
		t.Fatalf("expected second content to be *TextContent")
	}
	if msg.Text != prompt {
		t.Errorf("expected prompt text %q, got %q", prompt, msg.Text)
	}
}

func TestBuildSystemInstruction(t *testing.T) {
	personaMessage := "You are a helpful assistant."

	instructions := buildSystemInstruction(personaMessage)

	if len(instructions) != 2 {
		t.Fatalf("expected 2 instructions, got %d", len(instructions))
	}
	if instructions[0].Text != personaMessage {
		t.Errorf("expected first instruction to be persona message, got %q", instructions[0].Text)
	}
	if instructions[1].Text != inputHandlingInstruction {
		t.Errorf("expected second instruction to be inputHandlingInstruction, got %q", instructions[1].Text)
	}
}

func TestResolveContext_PlainString(t *testing.T) {
	stdinConsumed := false
	got, err := resolveContext("plain context string", "", &stdinConsumed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "<context>\nplain context string\n</context>"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestResolveContext_FromFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "notes.txt")
	fileContent := "some background notes"
	if err := os.WriteFile(tmpFile, []byte(fileContent), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	stdinConsumed := false
	got, err := resolveContext("@"+tmpFile, "", &stdinConsumed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "<context file=") {
		t.Errorf("expected result to contain '<context file=', got %q", got)
	}
	if !strings.Contains(got, fileContent) {
		t.Errorf("expected result to contain file content, got %q", got)
	}
}

func TestResolveContext_FileNotFound(t *testing.T) {
	stdinConsumed := false
	_, err := resolveContext("@/nonexistent/file.txt", "", &stdinConsumed)
	if err == nil {
		t.Error("expected error for nonexistent context file")
	}
}

func TestResolveContext_LabeledFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "notes.txt")
	if err := os.WriteFile(tmpFile, []byte("notes"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	stdinConsumed := false
	got, err := resolveContext("notes:@"+tmpFile, "", &stdinConsumed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, `file="`+tmpFile+`"`) {
		t.Errorf("expected result to contain file attribute, got %q", got)
	}
	if !strings.Contains(got, `name="notes"`) {
		t.Errorf("expected result to contain name attribute, got %q", got)
	}
}

func TestResolveContext_InlineTextWithColon_NotTreatedAsLabel(t *testing.T) {
	stdinConsumed := false
	raw := "https://example.com: see this"
	got, err := resolveContext(raw, "", &stdinConsumed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "<context>\n" + raw + "\n</context>"
	if got != want {
		t.Errorf("expected inline text with a colon to pass through untouched, got %q, want %q", got, want)
	}
}

func TestResolveContext_Stdin(t *testing.T) {
	stdinConsumed := false
	got, err := resolveContext("@-", "piped content", &stdinConsumed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "<context>\npiped content\n</context>"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
	if !stdinConsumed {
		t.Error("expected stdinConsumed to be true after resolving @-")
	}
}

func TestInputSizeWarning(t *testing.T) {
	if got := inputSizeWarning(100); got != "" {
		t.Errorf("expected no warning for small input, got %q", got)
	}
	if got := inputSizeWarning(warnInputBytes); got != "" {
		t.Errorf("expected no warning exactly at the threshold, got %q", got)
	}
	got := inputSizeWarning(warnInputBytes + 1)
	if got == "" {
		t.Error("expected a warning once the threshold is exceeded")
	}
	if !strings.Contains(got, "large") {
		t.Errorf("expected warning to mention the input is large, got %q", got)
	}
}

func TestFlagSource_OnlyOnce(t *testing.T) {
	require.True(t, flagSource.OnlyOnce,
		"flagSource must reject a second --source: exactly one source per prompt is the design")
}

// TestSourceFlag_OnlyOnce_RejectsDuplicate confirms our understanding of
// urfave/cli v3's OnlyOnce behavior using a throwaway flag definition
// (mirroring flagSource), rather than the shared package-level flagSource.
// urfave/cli tracks OnlyOnce via a counter on the *Flag itself that is
// never reset across cli.Command.Run() calls, so reusing flagSource here
// would make this test's second --source count against every other test
// in the package that also sets --source.
func TestSourceFlag_OnlyOnce_RejectsDuplicate(t *testing.T) {
	src := &cli.StringFlag{Name: "source", OnlyOnce: true}
	app := &cli.Command{
		Name:   "aico",
		Flags:  []cli.Flag{src},
		Action: func(ctx context.Context, cmd *cli.Command) error { return nil },
	}
	err := app.Run(context.Background(), []string{"aico", "--source", "a", "--source", "b"})
	require.Error(t, err)
}

func TestUserMessage_WithoutSource(t *testing.T) {
	prompt := "hello"

	userContents := make([]assistant.MessageContent, 0, 1)
	userContents = append(userContents, assistant.NewTextContent(prompt))
	userMsg := assistant.NewUserMessage(userContents...)

	contents := userMsg.GetContents()
	if len(contents) != 1 {
		t.Fatalf("expected 1 content, got %d", len(contents))
	}
	msg, ok := contents[0].(*assistant.TextContent)
	if !ok {
		t.Fatalf("expected content to be *TextContent")
	}
	if msg.Text != prompt {
		t.Errorf("expected prompt text %q, got %q", prompt, msg.Text)
	}
}
