package vstruct

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type MySecret struct {
	Host string `name:"host"           secret:"$ENV/mariadb"`
	Port string `name:"port"           secret:"$ENV/mariadb"`
	IP   string `name:"ip"             secret:"$ENV/mariadb"`
	User string `name:"root_user"      secret:"$ENV/mariadb"`
	Pass string `name:"root_pass"      secret:"$ENV/mariadb"`
	Foo  string `secret:"$ENV/mariadb"`
}

func setup(_ testing.TB) (func(t testing.TB, server *httptest.Server), *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`{
  "data": {
    "data": {
      "host": "maria",
      "port": "3306",
      "ip": "192.168.1.195",
      "root_user": "root",
      "root_pass": "root"
    }
  }
}
`))
	}))
	return func(tb testing.TB, server *httptest.Server) {
		server.Close()
	}, server
}

func TestNewFromHome(t *testing.T) {
	h, err := os.UserHomeDir()
	if err != nil {
		t.Log("no home directory found ", err.Error())
		t.SkipNow()
	}

	if _, err = os.Stat(h + "/.vault-token"); err == nil {
		if _, err = NewFromHome("", ""); err != nil {
			t.Error(err.Error())
		}
	} else {
		if _, err = NewFromHome("", ""); err == nil {
			t.Error("non-existing .vault-token should result in error")
		}
	}
}

func TestNewFromFileEmpty(t *testing.T) {
	if _, err := NewFromFile("", "", ""); err == nil {
		t.Error("empty file should return error")
	}
}

func TestNewFromFileInvalid(t *testing.T) {
	if _, err := NewFromFile("", "", "somefile"); err == nil {
		t.Error("non-existing file should return error")
	}
}

func TestNewFromFileExists(t *testing.T) {
	if _, err := NewFromFile("", "", "README.md"); err != nil {
		t.Error("existing file should not result in", err.Error())
	}
}

func TestReplacementVars(t *testing.T) {
	p, _ := NewFromFile("", "", "README.md")
	if p.rep == nil {
		t.Error("Replacements map should not be nil")
	}
	if len(p.rep) != 0 {
		t.Error("Replacements map should be of length 0")
	}
	p.Register("a", "b")
	if len(p.rep) != 1 {
		t.Error("Replacements map should be of length 1")
	}
	if p.rep["a"] != "b" {
		t.Error("Value of \"a\" should be \"b\"")
	}
}

func TestPointer(t *testing.T) {
	sec := MySecret{}
	p := New("", "", "")
	if err := p.Parse(sec); err == nil {
		t.Error("Parsing non pointer value should result in error")
	}
}

func TestStruct(t *testing.T) {
	sec := ""
	p := New("", "", "")
	if err := p.Parse(&sec); err == nil {
		t.Error("Parsing non pointer value should result in error")
	}
}

func TestParse(t *testing.T) {
	down, server := setup(t)
	defer down(t, server)

	p := New(server.URL, "kv", "something")
	p.Register("ENV", "staging")

	sec := new(MySecret)
	if err := p.Parse(sec); err != nil {
		t.Error("Parse should not result in error:", err.Error())
	}

	if sec.Host != "maria" {
		t.Errorf("Host: expected \"%s\"; got %s", "maria", sec.IP)
	}
	if sec.Port != "3306" {
		t.Errorf("IP: expected \"%s\"; got %s", "3306", sec.Port)
	}
	if sec.IP != "192.168.1.195" {
		t.Errorf("IP: expected \"%s\"; got %s", "192.168.1.195", sec.IP)
	}
	if sec.User != "root" {
		t.Errorf("IP: expected \"%s\"; got %s", "root", sec.User)
	}
	if sec.Pass != "root" {
		t.Errorf("IP: expected \"%s\"; got %s", "root", sec.Pass)
	}
}
