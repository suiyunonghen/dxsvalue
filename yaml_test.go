package dxsvalue

import (
	"fmt"
	"io/ioutil"
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

	vmapMerge = []byte(`
defaults: &defaults
  adapter:  postgres
  host:     localhost

development:
  database: myapp_development
  <<: *defaults
test: 234
<<: *defaults
`)
	vmaptest = []byte(`
languages:
 - Ruby
 - Perl
 - Python 
websites:
 YAML: yaml.org 
 Ruby: ruby-lang.org 
 Python: python.org 
 Perl: use.perl.org `)
)

func TestYamlParser(t *testing.T) {
	parser := newyamParser()
	defer freeyamlParser(parser)
	parser.parseData = vmaptest
	err := parser.parse()
	v := parser.root
	fmt.Println(parser.root.String())
	if err != nil{
		fmt.Println("解析yaml发生错误：",err)
		return
	}
	bt, _ := ioutil.ReadFile("./config.yml")
	v.LoadFromYaml(bt)
	fmt.Println(v.String())
}

func TestNewValueFromYaml(t *testing.T) {
	value,err := NewValueFromYaml(vmaptest,true)
	if err != nil{
		fmt.Println("发生错误：",err)
		return
	}
	fmt.Println(value.String())
	FreeValue(value)
}

