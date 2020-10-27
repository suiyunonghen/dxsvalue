package dxsvalue

import (
	"fmt"
	"testing"
)

var(
	vtest = []byte(`
-
 - bb: 
    gg: 32
    mm: asdf
 - aa
 -
  - cc
  - dd   
`)

	vtestB = []byte(`
- bb
- aa
-
  - cc
  - dd
`)
	vtestC = []byte(`
- bb
- aa
`)
	vmap = []byte(`
project:
  port: 8080
  name: &projectName epshealth-airobot-common

jwt:
  issuer: *projectName
  secret: c2VybmFtZSI6InRlc3QiLCJleHAiOjE2MDA5MjE0MjEsImlzcyI6ImVwc2hlYWx0aCIsIm5iZiI6MTYwMDkxNDIyMX0
  expires: 2h`)

	vmaparr = []byte(`
test:
  - 33
  - 44
  - 55
mm: sfasdf
istel: true`)

	vmapAdv = []byte(`
defaults: &defaults
  adapter:  postgres
  host:     localhost

development:
  database: myapp_development
  test: *defaults`)
)

func TestYamlParser(t *testing.T) {
	parser := newyamParser()
	defer freeyamlParser(parser)
	parser.fParsingValues = make([]yamlNode,0,10)
	parser.parseData = vmapAdv
	err := parser.parse()
	fmt.Println(parser.root.String())
	if err != nil{
		fmt.Println("解析yaml发生错误：",err)
	}
}