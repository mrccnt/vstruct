![lintcover workflow](https://github.com/mrccnt/vstruct/actions/workflows/lintcover.yml/badge.svg)
[![codecov](https://codecov.io/gh/mrccnt/vstruct/graph/badge.svg?token=INZR4MMMDZ)](https://codecov.io/gh/mrccnt/vstruct)

# Vault Struct Parser

Parser uses reflect and struct annotations to query a [Hashicorp Vault](https://www.vaultproject.io/) KV secrets engine
using vault HTTP API. Currently only `v2` of the engine is supported.

## Annotations

Two annotations are needed to define what you need. Both are mandatory to be able to query the KV storage. Properties
not having both annotations are getting ignored. 

| Annotation | Info                      |
|------------|---------------------------|
| `secret`   | Secret path behind engine |
| `name`     | The field name to read    |

Secrets are getting read only once, even if mentioned multiple times. As an additional feature you can register
replacement variables that can be used as placeholders in a `secret`.

## Example

Used vault instance has a secrets engine registered behind `kv`. Via vault CLI you would query those fields a little
something like:

```shell
vault kv get -field=root_pass kv/staging/mariadb
```

Usage with annotations:

```go
type MySecret struct {
	Host string `secret:"$ENV/mariadb" name:"host"`
	Port string `secret:"$ENV/mariadb" name:"port"`
	IP   string `secret:"$ENV/mariadb" name:"ip"`
	User string `secret:"$ENV/mariadb" name:"root_user"`
	Pass string `secret:"$ENV/mariadb" name:"root_pass"`
}

func main() {

	p := vstruct.New("https://vault.local:8200", "kv", "<token>")
	p.Register("ENV", "staging")

	sec := new(MySecret)
	if err := p.Parse(sec); err != nil {
		log.Fatalln(err)
	}
    
	fmt.Println(sec)
}

// Outputs:
// &{maria 3306 192.168.1.195 root root}
```

## About

This package should not encourage people to begin using vault as some kind of regular kv storage in their live/production
environments. Still, there are situations where this is just the right way to go. In my case I am frequently using it in
administrative CLI tools in combination with a strict policy management in vault (who can read what). Using a personal
vault token for that is more or less authentication plus authorization. It is also useful in the scope of continuous
integration where secrets are needed.

## Links

* [Hashicorp Vault](https://www.vaultproject.io/)
* [KV Secrets Engine HTTP API](https://developer.hashicorp.com/vault/api-docs/secret/kv/kv-v2)