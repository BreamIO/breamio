package config

import (
	"encoding/json"
	"strings"
	"testing"
)

type TestModule struct {
	conf *TestConfig
}

func (TestModule) String() string {
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

type ListTestModule struct {
	conf ListTestConfig
}

func (ListTestModule) String() string {
	return "ListModule"
}

func (t ListTestModule) Config() ConfigSection {
	return t.conf
}

type ListTestConfig []string

func TestLoad(t *testing.T) {
	tm := TestModule{conf: &TestConfig{}}
	ltm := ListTestModule{make(ListTestConfig, 0, 10)}

	sr := strings.NewReader(testConfigData)
	err := Load(sr)
	if err != nil {
		t.Fatal(err)
	}
	tc := config.Section(tm.String(), tm.Config()).(*TestConfig)
	if tc.A != 42 {
		t.Fail()
	}

	if tc.B != "Braxelibrax" {
		t.Fail()
	}

	ltc := config.Section(ltm.String(), &ltm.conf).(*ListTestConfig)
	t.Log(ltc)
	if len(*ltc) != 3 {
		t.Fail()
	}
}

func TestSection(t *testing.T) {
	tc := &TestConfig{1337, "goo", map[string][]string{"dar": []string{"tar", "var", "car"}}}
	tm := TestModule{conf: tc}
	tmp, _ := json.Marshal(tm.Config())
	config[tm.String()] = (*json.RawMessage)(&tmp)
	if Section(tm.String(), tm.Config()) != tc {
		t.Fail()
	}
}

func TestUpdate(t *testing.T) {
	tc := &TestConfig{1337, "goo", nil}
	tm := TestModule{conf: tc}
	Update(tm.String(), &TestConfig{1338, "doo", nil})
	tc2 := config.Section(tm.String(), tm.Config()).(*TestConfig)
	if tc2.A != 1338 {
		t.Fail()
	}
}

const testConfigData = `{
	"ListModule": ["Foo", "Bar", "Baz"],
	"TestModule": {
		"A": 42,
		"B": "Braxelibrax",

		"C": {
			"strings": [
				"hej då",
				"good bye"
				],
			"numbers": ["1", "2", "3"]
		}
	}
}`
