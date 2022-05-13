package fasthttp

import (
	"errors"
	"github.com/vela-security/vela-public/lua"
	"os"
)

var (
	defaultAccessJsonFormat = "[${time}] - [${remote_port}] - ${server_addr}:${server_port} ${remote_addr} " +
		"${method} [${scheme}] [${host}] ${uri} ${query} ${ua} ${referer} ${status} ${size} ${region_city}"
)

type config struct {
	//基础配置
	name        string
	listen      string
	network     string
	router      string
	handler     string
	keepalive   string
	reuseport   string
	daemon      string
	region      string
	notFound    string
	readTimeout int
	idleTimeout int

	//下面对象配置
	fd     *os.File
	output lua.Writer
	access func(*RequestCtx) []byte

	debug bool
}

func newConfig(L *lua.LState) *config {
	tab := L.CheckTable(1)
	cnn := &conversion{}
	cnn.pretreatment(defaultAccessJsonFormat)

	cfg := &config{
		readTimeout: 10,
		idleTimeout: 10,
		router:      xEnv.Prefix() + "/www/vhost",
		handler:     xEnv.Prefix() + "/www/handle",
		access:      cnn.Line,
	}

	tab.Range(func(key string, val lua.LValue) {
		switch key {
		case "name":
			cfg.name = val.String()
		case "daemon":
			cfg.daemon = val.String()
		case "listen":
			cfg.listen = val.String()
		case "network":
			cfg.network = val.String()
		case "reuseport":
			cfg.reuseport = val.String()
		case "keepalive":
			cfg.keepalive = val.String()

		case "read_timeout":
			cfg.readTimeout = lua.IsInt(val)

		case "idle_timeout":
			cfg.idleTimeout = lua.IsInt(val)

		case "output":
			cfg.output = checkOutputSdk(L, val)

		default:
			L.RaiseError("invalid web config %s field", key)
			return
		}
	})

	if e := cfg.verify(); e != nil {
		L.RaiseError("%v", e)
		return nil
	}
	return cfg
}

func (cfg *config) verify() error {
	if cfg.name == "" {
		return errors.New("invalid name")
	}

	return nil
}
