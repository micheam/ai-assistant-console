package assistant_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"

	assistantv1 "micheam.com/aico/internal/pb/assistant/v1"
)

func TestChatSession_Serialize(t *testing.T) {
	orig := &assistantv1.ChatSession{
		Id:        "sess-9D735EADE08A48E58C169E18E5AE3171",
		CreatedAt: timestamppb.Now(),
		SystemInstruction: &assistantv1.TextContent{
			Text: "This is a system instruction message.",
		},
		History: []*assistantv1.Message{
			{
				Role: assistantv1.Message_ROLE_USER,
				Contents: []*assistantv1.MessageContent{
					{
						Content: &assistantv1.MessageContent_Text{
							Text: &assistantv1.TextContent{
								Text: "Hello, how are you doing?",
							},
						},
					},
				},
			},
			{
				Role: assistantv1.Message_ROLE_ASSISTANT,
				Contents: []*assistantv1.MessageContent{
					{
						Content: &assistantv1.MessageContent_Text{
							Text: &assistantv1.TextContent{
								Text: "I am doing great, thank you!",
							},
						},
					},
				},
			},
		},
	}

	serialized, err := proto.Marshal(orig)
	if err != nil {
		t.Fatal(err)
	}

	var restored assistantv1.ChatSession
	if err := proto.Unmarshal(serialized, &restored); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(orig, &restored, protocmp.Transform()); diff != "" {
		t.Fatalf("unexpected diff (-want +got):\n%s", diff)
	}
}
