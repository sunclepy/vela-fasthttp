package fasthttp

import (
	"errors"
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/vela-security/vela-public/lua"
	region "github.com/vela-security/vela-region"
	"os"
	"path/filepath"
	"runtime/debug"
)

func checkHandleChains(L *lua.LState, seek int) *HandleChains {
	n := L.GetTop()
	if n < 2 {
		xEnv.Errorf("invalid args , #1 must string , #n+1 must be http.handler")
		return newHandleChains(0)
	}

	hc := newHandleChains(n - seek)
	var val lua.LValue
	for i := 2; i <= n; i++ {

		val = L.Get(i)
		switch val.Type() {

		//判断是否为加载
		case lua.LTString:
			hc.Store(val.String(), VHSTRING, i-2)

		case lua.LTObject:
			hd, ok := val.(*handle)
			if ok {
				hc.Store(hd, VHANDLER, i-2)
			} else {
				hc.Store(val.String(), VHSTRING, i-2)
			}

		case lua.LTFunction:
			hc.Store(val.(*lua.LFunction), VHFUNC, i-2)

		default:
			hc.Store(val.String(), VHSTRING, i-2)
		}
	}

	return hc
}

var (
	notFoundRouter = errors.New("not found router in co")
	invalidRouter  = errors.New("invalid router in co")
)

func checkRouter(L *lua.LState) (*vRouter, error) {
	if L.D == nil {
		return nil, notFoundRouter
	}

	r, ok := L.D.(*vRouter)
	if !ok {
		return nil, invalidRouter
	}

	return r, nil
}

func checkRequestCtx(L *lua.LState) *RequestCtx {
	if L.D == nil {
		L.RaiseError("invalid request context")
		return nil
	}

	ctx, ok := L.D.(*RequestCtx)
	if !ok {
		return nil
	}
	return ctx
}

func checkRegionSdk(L *lua.LState, val lua.LValue) *region.Region {

	switch val.Type() {
	case lua.LTNil:
		return nil

	case lua.LTProcData:
		r, ok := val.(*lua.ProcData).Data.(*region.Region)
		if !ok {
			L.RaiseError("invalid region sdk")
			return nil
		}
		return r

	default:
		//todo
	}

	L.RaiseError("invalid region object , got %s", val.Type().String())
	return nil
}

func checkOutputSdk(L *lua.LState, val lua.LValue) lua.Writer {
	switch val.Type() {
	case lua.LTNil:
		return nil
	case lua.LTProcData:
		w, ok := val.(*lua.ProcData).Data.(lua.Writer)
		if ok {
			return w
		}

	default:
		//todo
	}

	L.RaiseError("invalid output object , got %s", val.Type().String())
	return nil
}

func compileAccessFormat(format string, val string) func(ctx *RequestCtx) []byte {
	if len(val) == 0 {
		return nil
	}

	switch format {
	case "":
		return nil
	case "off":
		return nil
	case "json":
		cnn := &conversion{}
		cnn.pretreatment(val)
		return cnn.Json

	case "line":
		cnn := &conversion{}
		cnn.pretreatment(val)
		return cnn.Line

	default:
		return nil
	}
}

func compileHandle(filename string, args ...interface{}) (PoolItemIFace, error) {
	//重新获取
	co := xEnv.Coroutine()
	defer xEnv.Free(co)
	stat, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}

	err = xEnv.DoFile(co, filename)
	if err != nil {
		return nil, err
	}

	lv := co.Get(-1)
	if lv.Type() != lua.LTObject {
		return nil, fmt.Errorf("invalid handle type got %s", lv.Type().String())
	}

	hd, ok := lv.(*handle)
	if !ok {
		return nil, errors.New("invalid handle object")
	}

	hd.mtime = stat.ModTime().Unix()
	hd.name = filename
	xEnv.Errorf("handle %s compile succeed", filename)
	return hd, nil

}
func requireHandle(path, name string) (*handle, error) {
	filename := fmt.Sprintf("%s/%s.lua", path, name)

	//查看缓存
	item := handlePool.Get(filename)
	if item != nil {
		return item.val.(*handle), nil
	}

	hd, err := compileHandle(filename)
	if err != nil {
		return nil, err
	}

	handlePool.insert(filename, hd)
	return hd.(*handle), nil

}

func compileRouter(filename string, args ...interface{}) (PoolItemIFace, error) {

	//重新获取
	stat, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}
	var r *vRouter

	co := xEnv.Coroutine()
	defer xEnv.Free(co)

	//执行配置脚本
	err = xEnv.DoFile(co, filename)
	if err != nil {
		return nil, err
	}

	r, err = checkRouter(co)
	if err != nil {
		return nil, err
	}

	r.handler = args[0].(string)
	r.name = filename
	r.mtime = stat.ModTime().Unix()
	xEnv.Errorf("router %s compile succeed", filename)
	return r, nil

}

func requireRouter(path, handler, host string) (*vRouter, error) {
	filename := path + filepath.Join("/", host) + ".lua"

	//查看缓存
	item := routerPool.Get(filename)
	if item != nil {
		return item.val.(*vRouter), nil
	}

	r, err := compileRouter(filename, handler)
	if err != nil {
		return nil, err
	}

	routerPool.insert(filename, r)
	return r.(*vRouter), err
}

func checkLuaEof(ctx *RequestCtx) bool {
	uv := ctx.UserValue(eof_uv_key)
	if uv == nil {
		return false
	}

	v, ok := uv.(bool)
	if !ok {
		return false
	}

	return v
}

func panicHandler(ctx *RequestCtx, val interface{}) {
	ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
	e := fmt.Sprintf("%v %s", val, debug.Stack())
	ctx.Response.SetBodyString(e)
}

func compileHandleBody(data string) func(*RequestCtx) error {
	// helo ${host} ${uri} ${param_name}

	cnn := &conversion{}
	cnn.pretreatment(data)

	return func(ctx *RequestCtx) error {
		cnn.Response(ctx)
		return nil
	}
}
