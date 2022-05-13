package fasthttp

import "github.com/vela-security/vela-public/lua"

func fsRedirect(L *lua.LState) int {
	n := L.GetTop()
	var path string
	var code int

	switch n {
	case 1:
		path = L.CheckString(1)
		code = 302
	case 2:
		path = L.CheckString(1)
		code = L.CheckInt(2)
	default:
		return 0
	}

	ctx := checkRequestCtx(L)
	ctx.Redirect(path, code)
	return 0
}

func rqhL(L *lua.LState) int {
	return fsHeaderHelper(L, false)
}

func rphL(L *lua.LState) int {
	return fsHeaderHelper(L, true)
}

func (fsc *fsContext) Index(co *lua.LState, key string) lua.LValue {
	ctx := checkRequestCtx(co)
	switch key {
	case "json":
		return fsc.sayJson
	case "say":
		return fsc.say
	case "append":
		return fsc.append
	case "exit":
		return fsc.exit
	case "eof":
		return fsc.eof
	case "redirect":
		return fsc.rdt
	case "req_header":
		return fsc.rqh
	case "resp_header":
		return fsc.rph
	case "try":
		return fsc.try
	case "bind":
		return fsc.bind
	}

	return k2v(ctx, key)
}

func (fsc *fsContext) NewIndex(co *lua.LState, key string, val lua.LValue) {
	ctx := checkRequestCtx(co)
	if key == "path" {
		ctx.URI().SetPath(val.String())
	}
}
