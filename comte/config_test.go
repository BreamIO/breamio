package comte

import (
	"strings"
	"testing"
)

type TestModule struct {
	conf *TestConfig
}

func (TestModule) Name() string {
	return "TestModule"
}

func (t TestModule) Config() ConfigSection {
	return t.conf
}

type TestConfig struct {
	A int
	B string
	C map[string][]string
}

func setup(c Configuration) {
	tm := TestModule{conf: &TestConfig{}}
	c[tm.Name()] = tm.Config()
}

func TestRegister(t *testing.T) {
	tm := TestModule{conf: &TestConfig{}}
	Register(tm)

	if v, ok := config[tm.Name()]; ok {
		if v != tm.Config() {
			t.Fail()
		}
	} else {
		t.Fail()
	}
}

func TestLoad(t *testing.T) {
	setup(config)

	sr := strings.NewReader(testConfigData)
	err := Load(sr)
	if err != nil {
		t.Fatal(err)
	}
	tc := config[TestModule{}.Name()].(*TestConfig)
	t.Log(tc)
	t.Log(config)
	if tc.A != 42 {
		t.Fail()
	}

	if tc.B != "Braxelibrax" {
		t.Fail()
	}
}

func TestSection(t *testing.T) {

}

const testConfigData = `{
	    "TestModule": {
	        "A": 42,
	        "B": "Braxelibrax",
	        "C": {
	            "strings": [
	                "hej då",
	                "good bye"
	            ],
	            "numbers": [
	                "1",
	                "2",
	                "3"
	            ]
	        }
	    }
}`
