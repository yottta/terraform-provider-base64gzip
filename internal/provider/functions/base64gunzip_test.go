package functions

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBase64GunzipFunction_Metadata(t *testing.T) {
	t.Parallel()

	if got := metadataName(context.Background(), Base64GunzipFunction{}); got != "base64gunzip" {
		t.Errorf("expected the function to be named %q, got %q", "base64gunzip", got)
	}
}

func TestBase64GunzipFunction_Definition(t *testing.T) {
	t.Parallel()

	assertSingleStringParameterDefinition(t, definitionOf(context.Background(), Base64GunzipFunction{}))
}

// TestBase64GunzipFunction_Run_Golden pins the decoding of the hardcoded encoded values. See
// goldenFixtures for why these values are hardcoded.
func TestBase64GunzipFunction_Run_Golden(t *testing.T) {
	t.Parallel()

	for _, tt := range goldenFixtures {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := runFunction(t, Base64GunzipFunction{}, types.StringValue(tt.encoded))
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if result.IsNull() || result.IsUnknown() {
				t.Fatalf("expected a known, non null result, got %s", result)
			}

			if result.ValueString() != tt.plain {
				t.Errorf("unexpected decoding of %q:\n\texpected: %q\n\tgot:      %q", tt.encoded, tt.plain, result.ValueString())
			}
		})
	}
}

// TestBase64GunzipFunction_Run_Errors covers every failure path of the function.
func TestBase64GunzipFunction_Run_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		errPrefix string
	}{
		{
			name:      "not base64",
			input:     "this is definitely not base64!",
			errPrefix: "failed to decode base64 data: ",
		},
		{
			name:      "base64 with bad padding",
			input:     "H4sIAAAAAAAA/wAEAPv/dGVzdAAAAP//AwAMfn/YBAAAAA=",
			errPrefix: "failed to decode base64 data: ",
		},
		{
			name:      "empty string",
			input:     "",
			errPrefix: "failed to read gzip data: ",
		},
		{
			name:      "valid base64 but not gzip",
			input:     base64.StdEncoding.EncodeToString([]byte("not a gzip stream")),
			errPrefix: "failed to read gzip data: ",
		},
		{
			name:      "truncated gzip header",
			input:     base64.StdEncoding.EncodeToString(mustDecodeBase64(t, goldenFixtures[1].encoded)[:5]),
			errPrefix: "failed to read gzip data: ",
		},
		{
			name:      "truncated gzip body",
			input:     base64.StdEncoding.EncodeToString(truncate(mustDecodeBase64(t, goldenFixtures[5].encoded), 8)),
			errPrefix: "failed to decompress gzip data: ",
		},
		{
			name:      "corrupted gzip payload",
			input:     base64.StdEncoding.EncodeToString(corruptByte(mustDecodeBase64(t, goldenFixtures[5].encoded), 14)),
			errPrefix: "failed to decompress gzip data: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := runFunction(t, Base64GunzipFunction{}, types.StringValue(tt.input))
			if err == nil {
				t.Fatalf("expected an error for input %q", tt.input)
			}

			if !strings.HasPrefix(err.Text, tt.errPrefix) {
				t.Errorf("expected the error to start with %q, got %q", tt.errPrefix, err.Text)
			}

			if err.FunctionArgument == nil {
				t.Fatal("expected the error to point at a function argument")
			}
			if *err.FunctionArgument != 0 {
				t.Errorf("expected the error to point at argument 0, got %d", *err.FunctionArgument)
			}
		})
	}
}

// TestBase64GunzipFunction_Run_MissingArgument covers the argument retrieval failure path.
func TestBase64GunzipFunction_Run_MissingArgument(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	resp := function.RunResponse{
		Result: function.NewResultData(types.StringUnknown()),
	}

	Base64GunzipFunction{}.Run(ctx, function.RunRequest{Arguments: function.NewArgumentsData(nil)}, &resp)

	if resp.Error == nil {
		t.Fatal("expected an error when no argument data is provided")
	}
}

func mustDecodeBase64(t *testing.T, str string) []byte {
	t.Helper()

	decoded, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		t.Fatalf("failed to decode the test fixture %q: %s", str, err)
	}

	return decoded
}

// truncate removes the last n bytes of the given slice.
func truncate(data []byte, n int) []byte {
	if n >= len(data) {
		return nil
	}

	return data[:len(data)-n]
}

// corruptByte flips all the bits of the byte at the given index.
func corruptByte(data []byte, index int) []byte {
	corrupted := make([]byte, len(data))
	copy(corrupted, data)
	corrupted[index] ^= 0xFF

	return corrupted
}
