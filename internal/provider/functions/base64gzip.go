package functions

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"

	"github.com/hashicorp/terraform-plugin-framework/function"
)

var (
	_ function.Function = &Base64GzipFunction{}
)

type Base64GzipFunction struct {
}

func (b Base64GzipFunction) Metadata(_ context.Context, _ function.MetadataRequest, response *function.MetadataResponse) {
	response.Name = "base64gzip"
}

func (b Base64GzipFunction) Definition(_ context.Context, _ function.DefinitionRequest, response *function.DefinitionResponse) {
	response.Definition = function.Definition{
		Summary:     "Compresses a string with gzip and encodes the result in Base64",
		Description: "Compresses the given string with gzip and then encodes the result to Base64 using the standard alphabet, matching the behaviour of the built-in `base64gzip` function.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "str",
				Description: "The string to compress and encode.",
			},
		},
		Return: function.StringReturn{},
	}
}

func (b Base64GzipFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var str string

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &str))
	if resp.Error != nil {
		return
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	if _, err := gz.Write([]byte(str)); err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, "failed to write gzip raw data: "+err.Error()))
		return
	}
	if err := gz.Flush(); err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, "failed to flush gzip writer: "+err.Error()))
		return
	}
	if err := gz.Close(); err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, "failed to close gzip writer: "+err.Error()))
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, base64.StdEncoding.EncodeToString(buf.Bytes())))
}
