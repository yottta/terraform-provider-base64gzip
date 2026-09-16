package functions

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// goldenFixtures holds the canonical mapping between a plain string and its gzip+base64 encoded form.
//
// These values are hardcoded on purpose: they pin the exact byte-for-byte output of the provider
// functions so that it stays compatible with OpenTofu's built-in base64gzip/base64gunzip functions.
// A failure here means the compress/flate output changed (typically after a Go toolchain upgrade)
// and that the encoded values produced by this provider would no longer match the ones produced by
// OpenTofu core.
var goldenFixtures = []struct {
	name    string
	plain   string
	encoded string
}{
	{
		name:    "empty string",
		plain:   "",
		encoded: "H4sIAAAAAAAA/wAAAP//AwAAAAAAAAAAAA==",
	},
	{
		name:    "short ascii",
		plain:   "test",
		encoded: "H4sIAAAAAAAA/wAEAPv/dGVzdAAAAP//AwAMfn/YBAAAAA==",
	},
	{
		name:    "punctuation",
		plain:   "Hello, World!",
		encoded: "H4sIAAAAAAAA/wANAPL/SGVsbG8sIFdvcmxkIQAAAP//AwDQw0rsDQAAAA==",
	},
	{
		name:    "multi byte utf8",
		plain:   "héllo wörld é世界",
		encoded: "H4sIAAAAAAAA/wAWAOn/aMOpbGxvIHfDtnJsZCDDqeS4lueVjAAAAP//AwCmau0OFgAAAA==",
	},
	{
		name:    "multiline",
		plain:   "line1\nline2\n",
		encoded: "H4sIAAAAAAAA/wAMAPP/bGluZTEKbGluZTIKAAAA//8DAFddjToMAAAA",
	},
	{
		name:    "highly compressible",
		plain:   strings.Repeat("a", 1024),
		encoded: "H4sIAAAAAAAA/0ocBaNgFIxYAAAAAP//AwC5l1V8AAQAAA==",
	},
}

// runFunction invokes the Run method of the given function with the given positional arguments and
// returns the string result together with any function error that was raised.
func runFunction(t *testing.T, fn function.Function, args ...attr.Value) (types.String, *function.FuncError) {
	t.Helper()

	ctx := context.Background()

	req := function.RunRequest{
		Arguments: function.NewArgumentsData(args),
	}
	resp := function.RunResponse{
		Result: function.NewResultData(types.StringUnknown()),
	}

	fn.Run(ctx, req, &resp)

	if resp.Error != nil {
		return types.StringUnknown(), resp.Error
	}

	result, ok := resp.Result.Value().(types.String)
	if !ok {
		t.Fatalf("expected the result to be a types.String, got %T", resp.Result.Value())
	}

	return result, nil
}

// metadataName returns the name the given function registers itself under.
func metadataName(ctx context.Context, fn function.Function) string {
	resp := function.MetadataResponse{}
	fn.Metadata(ctx, function.MetadataRequest{}, &resp)

	return resp.Name
}

// definitionOf returns the definition of the given function.
func definitionOf(ctx context.Context, fn function.Function) function.Definition {
	resp := function.DefinitionResponse{}
	fn.Definition(ctx, function.DefinitionRequest{}, &resp)

	return resp.Definition
}

// assertSingleStringParameterDefinition validates the parts of a definition that are shared between
// both functions of this provider: a single required string parameter and a string return value.
func assertSingleStringParameterDefinition(t *testing.T, def function.Definition) {
	t.Helper()

	if def.Summary == "" {
		t.Error("expected the definition to have a summary")
	}
	if def.Description == "" {
		t.Error("expected the definition to have a description")
	}

	if len(def.Parameters) != 1 {
		t.Fatalf("expected exactly 1 parameter, got %d", len(def.Parameters))
	}

	param, ok := def.Parameters[0].(function.StringParameter)
	if !ok {
		t.Fatalf("expected the parameter to be a function.StringParameter, got %T", def.Parameters[0])
	}
	if param.Name != "str" {
		t.Errorf("expected the parameter to be named %q, got %q", "str", param.Name)
	}
	if param.Description == "" {
		t.Error("expected the parameter to have a description")
	}

	if def.VariadicParameter != nil {
		t.Errorf("expected no variadic parameter, got %T", def.VariadicParameter)
	}

	if _, ok := def.Return.(function.StringReturn); !ok {
		t.Errorf("expected the return to be a function.StringReturn, got %T", def.Return)
	}
}

// TestRoundTrip asserts that base64gunzip is the exact inverse of base64gzip.
func TestRoundTrip(t *testing.T) {
	t.Parallel()

	for _, tt := range goldenFixtures {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			encoded, err := runFunction(t, Base64GzipFunction{}, types.StringValue(tt.plain))
			if err != nil {
				t.Fatalf("unexpected error while compressing: %s", err)
			}

			decoded, err := runFunction(t, Base64GunzipFunction{}, encoded)
			if err != nil {
				t.Fatalf("unexpected error while decompressing: %s", err)
			}

			if decoded.ValueString() != tt.plain {
				t.Errorf("expected the round trip to return %q, got %q", tt.plain, decoded.ValueString())
			}
		})
	}
}
