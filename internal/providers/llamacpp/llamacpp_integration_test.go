//go:build integration

package llamacpp

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"micheam.com/aico/internal/complete"
)

// envTestEndpoint names the environment variable holding the /infill URL of
// a running llama-server (e.g. http://127.0.0.1:8012/infill). The test is
// skipped when it is unset.
const envTestEndpoint = "AICO_TEST_LLAMACPP_ENDPOINT"

func TestIntegration_Complete(t *testing.T) {
	endpoint := os.Getenv(envTestEndpoint)
	if endpoint == "" {
		t.Skipf("%s is not set", envTestEndpoint)
	}
	c := New(Options{
		Endpoint:      endpoint,
		NPredict:      16,
		TMaxPromptMS:  5000,
		TMaxPredictMS: 5000,
	})
	req := complete.Request{
		Prefix: "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.",
		Suffix: "\n}\n",
		Indent: 1,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	first, err := c.Complete(ctx, req)
	require.NoError(t, err)
	t.Logf("first:  %+v", first)

	// The same request again must reuse the prompt cache filled by the
	// first one, since every request is sent with cache_prompt.
	req.Count = 2
	second, err := c.Complete(ctx, req)
	require.NoError(t, err)
	t.Logf("second: %+v", second)
	assert.Positive(t, second.Timings.CachedTokens)
	assert.Len(t, second.Candidates, 2)
}
