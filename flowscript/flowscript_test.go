package flowscript

import (
	"strings"
	"testing"
)

func TestEvaluateScript(t *testing.T) {
	ge := createTestGlobalEnvironment()
	{
		value, err := EvaluateScript("hoge + 123 + \"foo\"", ge)
		if err != nil {
			t.Fatalf("error: %s", err)
		}
		if str, ok := value.(StringValue); !ok || str.Value() != "hoge123foo" {
			t.Fatalf("bad result: %s", value)
		}
	}
	{
		_, err := EvaluateScript("hoge + !", ge)
		if err == nil || !strings.HasPrefix(err.Error(), "parse error") {
			t.Fatalf("error: %s", err)
		}
	}
}
