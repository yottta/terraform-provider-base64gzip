package functions

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"io"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBase64GzipFunction_Metadata(t *testing.T) {
	t.Parallel()

	if got := metadataName(context.Background(), Base64GzipFunction{}); got != "base64gzip" {
		t.Errorf("expected the function to be named %q, got %q", "base64gzip", got)
	}
}

func TestBase64GzipFunction_Definition(t *testing.T) {
	t.Parallel()

	assertSingleStringParameterDefinition(t, definitionOf(context.Background(), Base64GzipFunction{}))
}

// TestBase64GzipFunction_Run_Golden pins the exact output of the function. See goldenFixtures for
// why these values are hardcoded.
func TestBase64GzipFunction_Run_Golden(t *testing.T) {
	t.Parallel()

	for _, tt := range goldenFixtures {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := runFunction(t, Base64GzipFunction{}, types.StringValue(tt.plain))
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if result.IsNull() || result.IsUnknown() {
				t.Fatalf("expected a known, non null result, got %s", result)
			}

			if result.ValueString() != tt.encoded {
				t.Errorf("unexpected encoding of %q:\n\texpected: %q\n\tgot:      %q", tt.plain, tt.encoded, result.ValueString())
			}
		})
	}
}

// TestBase64GzipFunction_Run_IsValidGzip asserts that the output of the function is a valid base64
// encoded gzip stream that decompresses back into the given input, independently of the hardcoded
// golden values.
func TestBase64GzipFunction_Run_IsValidGzip(t *testing.T) {
	t.Parallel()

	for _, tt := range goldenFixtures {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := runFunction(t, Base64GzipFunction{}, types.StringValue(tt.plain))
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			compressed, decodeErr := base64.StdEncoding.DecodeString(result.ValueString())
			if decodeErr != nil {
				t.Fatalf("expected a valid base64 output, got %q: %s", result.ValueString(), decodeErr)
			}

			gz, gzErr := gzip.NewReader(bytes.NewReader(compressed))
			if gzErr != nil {
				t.Fatalf("expected a valid gzip stream: %s", gzErr)
			}
			t.Cleanup(func() { _ = gz.Close() })

			decompressed, readErr := io.ReadAll(gz)
			if readErr != nil {
				t.Fatalf("failed to decompress the output: %s", readErr)
			}

			if string(decompressed) != tt.plain {
				t.Errorf("expected the decompressed output to be %q, got %q", tt.plain, string(decompressed))
			}
		})
	}
}

// TestBase64GzipFunction_Run_LargeInput makes sure inputs that exceed the internal deflate buffers
// are handled and that compression actually kicks in for repetitive data.
func TestBase64GzipFunction_Run_LargeInput(t *testing.T) {
	t.Parallel()

	input := strings.Repeat("the quick brown fox jumps over the lazy dog\n", 10_000)

	result, err := runFunction(t, Base64GzipFunction{}, types.StringValue(input))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(result.ValueString()) >= len(input) {
		t.Errorf("expected the compressed output (%d bytes) to be smaller than the input (%d bytes)", len(result.ValueString()), len(input))
	}

	decoded, err := runFunction(t, Base64GunzipFunction{}, result)
	if err != nil {
		t.Fatalf("unexpected error while decompressing: %s", err)
	}

	if decoded.ValueString() != input {
		t.Error("expected the large input to survive a round trip")
	}
}

// TestBase64GzipFunction_Run_MissingArgument covers the argument retrieval failure path.
func TestBase64GzipFunction_Run_MissingArgument(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	resp := function.RunResponse{
		Result: function.NewResultData(types.StringUnknown()),
	}

	Base64GzipFunction{}.Run(ctx, function.RunRequest{Arguments: function.NewArgumentsData(nil)}, &resp)

	if resp.Error == nil {
		t.Fatal("expected an error when no argument data is provided")
	}
}
