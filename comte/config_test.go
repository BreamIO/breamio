package comte

import (
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

func TestRegister(t *testing.T) {
	tm := TestModule{conf: &TestConfig{}}
	Register(tm)

	if v, ok := config[tm.String()]; ok {
		if v != tm.Config() {
			t.Fail()
		}
	} else {
		t.Fail()
	}
}

func TestLoad(t *testing.T) {
	tm := TestModule{conf: &TestConfig{}}
	config[tm.String()] = tm.Config()
	ltm := ListTestModule{make(ListTestConfig, 0, 10)}
	config[ltm.String()] = ltm.Config()

	sr := strings.NewReader(testConfigData)
	err := Load(sr)
	if err != nil {
		t.Fatal(err)
	}
	tc := config[TestModule{}.String()].(*TestConfig)
	t.Log(tc)
	t.Log(config)
	if tc.A != 42 {
		t.Fail()
	}

	if tc.B != "Braxelibrax" {
		t.Fail()
	}

	ltc := config[ListTestModule{}.String()].(ListTestConfig)
	t.Log(ltc)
	if len(ltc) != 3 {
		t.Fail()
	}
}

func TestSection(t *testing.T) {
	tc := &TestConfig{1337, "goo", map[string][]string{"dar": []string{"tar", "var", "car"}}}
	tm := TestModule{conf: tc}
	config[tm.String()] = tm.Config()
	if Section(tm.String()) != tc {
		t.Fail()
	}
}

func TestUpdate(t *testing.T) {
	tc := &TestConfig{1337, "goo", nil}
	tm := TestModule{conf: tc}
	config[tm.String()] = tm.Config()
	Update(tm.String(), &TestConfig{1338, "doo", nil})
	tc2 := config[tm.String()].(*TestConfig)
	if tc2.A != 1338 {
		t.Fail()
	}
}

const testConfigData = `
ListModule = ["Foo", "Bar", "Baz"]

[TestModule]
A = 42
B = "Braxelibrax"

[TestModule.C]
strings = [
	"hej då",
	"good bye"
	]
numbers = ["1", "2", "3"]
`
