package vstruct

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"
)

const (
	apiVersion  = "v1"
	segData     = "data"
	tokenFile   = ".vault-token"
	vaultHeader = "X-Vault-Token"
	httpTimeout = 10 * time.Second
	tagSec      = "secret"
	tagName     = "name"
	op          = "$"
)

type Parser struct {
	Client  http.Client
	baseURL string
	token   string
	mount   string
	rep     map[string]string
}

func NewFromHome(baseURL, mount string) (*Parser, error) {
	h, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return NewFromFile(baseURL, mount, fmt.Sprintf("%s/%s", h, tokenFile))
}

func NewFromFile(baseURL, mount, file string) (*Parser, error) {
	bs, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	return New(baseURL, mount, strings.TrimSpace(string(bs))), nil

}

func New(baseURL, enginePath, token string) *Parser {
	return &Parser{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		token:   token,
		mount:   enginePath,
		Client:  http.Client{Timeout: httpTimeout},
		rep:     map[string]string{},
	}
}

func (p *Parser) Register(key, value string) {
	p.rep[key] = value
}

func (p *Parser) Parse(f interface{}) error {
	v := reflect.ValueOf(f)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("not a pointer; is %T", f)
	}
	e := v.Elem()
	if e.Kind() != reflect.Struct {
		return fmt.Errorf("not a struct; is %T", f)
	}
	store := map[string]map[string]string{}
	t := e.Type()
	for x := 0; x < t.NumField(); x++ {
		if t.Field(x).IsExported() {
			sec, ok1 := t.Field(x).Tag.Lookup(tagSec)
			name, ok2 := t.Field(x).Tag.Lookup(tagName)
			if !ok1 || !ok2 {
				continue
			}
			for s, r := range p.rep {
				sec = strings.ReplaceAll(sec, op+s, r)
			}
			if _, ok := store[sec]; !ok {
				m, err := p.read(sec)
				if err != nil {
					return err
				}
				store[sec] = m
			}
			if secval, exists := store[sec][name]; exists {
				e.Field(x).SetString(secval)
			}
		}
	}
	return nil
}

func (p *Parser) read(securl string) (map[string]string, error) {

	var req *http.Request
	var res *http.Response
	var err error

	securl = fmt.Sprintf("%s/%s/%s/%s/%s", p.baseURL, apiVersion, p.mount, segData, strings.Trim(securl, "/"))

	if req, err = http.NewRequest(http.MethodGet, securl, nil); err != nil {
		return nil, err
	}

	req.Header.Set(vaultHeader, p.token)

	if res, err = p.Client.Do(req); err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New("status error: " + res.Status)
	}

	var bs []byte
	if bs, err = io.ReadAll(res.Body); err != nil {
		return nil, err
	}

	obj := struct {
		Data struct {
			Data map[string]string `json:"data"`
		} `json:"data"`
	}{}

	if err = json.Unmarshal(bs, &obj); err != nil {
		return nil, err
	}

	return obj.Data.Data, nil
}
