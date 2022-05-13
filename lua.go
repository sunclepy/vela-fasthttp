package fasthttp

import (
	"github.com/vela-security/vela-public/assert"
	"github.com/vela-security/vela-public/lua"
)

func newLuaServer(L *lua.LState) int {
	cfg := newConfig(L)
	proc := L.NewProc(cfg.name, typeof)
	if proc.IsNil() {
		proc.Set(newServer(cfg))
	} else {
		proc.Data.(*server).cfg = cfg
	}

	L.Push(proc)
	return 1
}

func WithEnv(env assert.Environment) {
	xEnv = env

	fs := lua.NewUserKV()
	fs.Set("context", newContext())
	fs.Set("new", lua.NewFunction(newLuaServer))
	fs.Set("h", lua.NewFunction(newLuaHandle))
	fs.Set("handle", lua.NewFunction(newLuaHandle))
	fs.Set("router", lua.NewFunction(newLuaRouter))
	fs.Set("filter", lua.NewFunction(newLuaFilter))
	fs.Set("header", lua.NewFunction(newLuaHeader))
	fs.Set("vhost", lua.NewFunction(newLuaHost))
	env.Global("web", fs)
}
