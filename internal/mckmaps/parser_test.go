package mckmaps

import (
	"fmt"
	"testing"

	"github.com/kumasuke120/mockuma/internal/myjson"
)

const mappingsJson1 = `
{
  "@type": "mappings",
  "mappings": [
    {
      "uri": "/",
      "method": "GET",
      "policies": {
        "returns": {
          "latency": [
            100,
            3000
          ],
          "headers": {
            "Content-Type": "application/json; charset=utf8"
          },
          "body": "abc123"
        }
      }
    }
  ]
}`

func TestMappingsParser_parse(t *testing.T) {
	j1, e1 := myjson.Unmarshal([]byte(mappingsJson1))
	if e1 != nil {
		t.Fatal("j1: json parse failed")
	}
	m1 := &mappingsParser{json: j1}
	p1, e1 := m1.parse()
	if e1 != nil {
		t.Error("m1: parse failed")
	}
	fmt.Println(p1)
}
