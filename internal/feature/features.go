package feature

import (
	"os"
	"strconv"
)

var DOCUMENT_CODE_BLOCK_DEFAULT_INTERACTIVE = true

func init() {
	if val, err := strconv.ParseBool(os.Getenv("RUNME_FLAGS_DOCUMENT_CODE_BLOCK_DEFAULT_INTERACTIVE")); err == nil {
		DOCUMENT_CODE_BLOCK_DEFAULT_INTERACTIVE = val
	}
}
