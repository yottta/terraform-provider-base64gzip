package functions

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"io"

	"github.com/hashicorp/terraform-plugin-framework/function"
)

var (
	_ function.Function = &Base64GunzipFunction{}
)

// Base64GunzipFunction is a provider function that receives a base64 gzipped input and returns the undecoded and uncompressed
// result.
type Base64GunzipFunction struct {
}

func (b Base64GunzipFunction) Metadata(_ context.Context, _ function.MetadataRequest, response *function.MetadataResponse) {
	response.Name = "base64gunzip"
}

func (b Base64GunzipFunction) Definition(_ context.Context, _ function.DefinitionRequest, response *function.DefinitionResponse) {
	response.Definition = function.Definition{
		Summary:     "Decodes a Base64 string and decompresses it with gzip",
		Description: "Decodes the given Base64 string using the standard alphabet and decompresses the result with gzip. This is the inverse of `base64gzip`.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "str",
				Description: "The Base64 encoded, gzip compressed string to decode.",
			},
		},
		Return: function.StringReturn{},
	}
}

// Run implements the base64 decoding and gzip decompression of the input.
// This follows the implementation of the OpenTofu's core function: https://github.com/opentofu/opentofu/blob/8368dc8f09d8b2863b88d6373c1076f548ac638d/internal/lang/funcs/encoding.go#L195-L213
func (b Base64GunzipFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var str string

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &str))
	if resp.Error != nil {
		return
	}

	compressed, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, "failed to decode base64 data: "+err.Error()))
		return
	}

	gz, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, "failed to read gzip data: "+err.Error()))
		return
	}
	defer func() { _ = gz.Close() }()

	gunzip, err := io.ReadAll(gz)
	if err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, "failed to decompress gzip data: "+err.Error()))
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, string(gunzip)))
}
