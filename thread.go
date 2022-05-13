package fasthttp

import "github.com/vela-security/vela-public/lua"

func newLuaThread(ctx *RequestCtx) *lua.LState {
	uv := ctx.UserValue(thread_uv_key)
	if uv != nil {
		return uv.(*lua.LState)
	}

	//设置ctx
	co := xEnv.Coroutine()
	co.D = ctx
	ctx.SetUserValue(thread_uv_key, co)

	return co
}

func freeLuaThread(ctx *RequestCtx) {
	co := ctx.UserValue(thread_uv_key)
	if co == nil {
		return
	}

	xEnv.Free(co.(*lua.LState))
}
