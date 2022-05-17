# vela-fasthttp
主要是用的web和lua的开发框架 可以灵活的处理业务 利用lua的热更新和加载的原理

## proc = web.new{conf}
- 按照配置文件启动一个web服务
- 返回对象是一个proc data object

### 字段
- name
- bind
- keepalive
- reuseport
- output

### 函数
- vhost(hostname , cfg)
- start()
- format(codec , string)
- addr(string)
- to(lua.write)

### 样例
```lua
    local http = web.new{
        -- proc 服务名称 
        name = "demo_web", 
        -- 监听端口
        bind = "tcp://0.0.0.0:9090?keepalive=on&read_timeout=100&idle_timeout=100", --监听的端口
        -- 默认没有发现的放回结果 
        not_found = "not_found",
    }

    local r = http.vhost("x.vela.com" , {})
    r.GET("/aa" , web.handle"xxxxxxx")
    r.GET("/cc" , function() end)

    http.start() 
```

## web.router{cfg} 
- 利用的web的router路由逻辑，完成默认路由的注入 ， 利用web的快速匹配模式完成路由查找
- 下面是www.a.com的主机的配置,文件路径:resource/web/server.d/www.a.com.lua

```lua
    local r = web.router{
    -- 日志格式 , 关闭：off 
    access_format = "time,server_addr,server_port,remote_addr,host,path,query,ua,referer,status,content-lenght,region_info,http_risk",

    -- 日志的格式 
    access_encode = "json",

    -- 记录IP的位置信息
    access_region = "x-real-ip",
    
    --
    output = vela.file{},
    
    interceptor = function()
        if ctx.status == 500 then
            ctx.say([[{"code": 500 , "message":"invalid request"}]])    
        end
    end
}
```

完整的例子 
```lua
local ctx = web.context-- 用户请求周期变量
local json = vela.json
local function auth() 
    local u = json.decode(ctx.body_raw)
    if u.name == "admin" and u.pass == "123654" then
        ctx.say(json.encode({code=200 , message= u.name .. "login success"}))
        ctx.exit(200) 
        return
    end
    ctx.say(json.encode({code=200 , message= "login fail"}))
    ctx.exit(200)
end

r.POST("/login" , auth) --注册路由

r.GET("/info" , 
    web.handle{
        code = 200,
        filter = web.filter{
           "$host == www.a.com,www.b.com,www.c.com",
           "$uri == /info"
        },
        
        header = web.header{
            ['server'] = "vela-web-test-v1.0", 
        },
        
        body = "helo",
        hook = function()
            ctx.say("helo by hook")
        end,
        eof = true 
    },
    "defaut_handle" --这个是查找公共库下的handle处理逻辑 ,一般为default_handle.lua
)
```
### 函数说明 
#### 1.路由配置
- router.GET
- router.HEAD
- router.POST
- router.PUT
- router.PATCH
- router.DELETE
- router.CONNECT
- router.OPTIONS
- router.TRACE
- router.POST
- router.ANY 忽略发方法名 , 注意: router.ANY("*" , web.handle...)
  
- 语法:  r.METHOD(path string , web.handle ... )
- 参数 path： 代表路径的 完全兼容 web.router的路径语法 如:/api/{name}/{val:*}
- 参数 handle: 就是用web.handle构造的对象

#### 2. ctx周期变量的使用
- web.context.say_json
自动encode obj 对象 并且发送JSON对象 , obj 需要满足是userdata anydata 且满足ToJson 接口
```lua
    local ctx = web.context
    ctx.say_json(obj)
```
- web.context.bind_json
自动绑定请求参数 如果异常会自动退出请求
```lua
    local ctx = web.context
    local v = ctx.bind_json()
    ctx.say(v.int("id"))
    ctx.say(v.str("name"))
    ctx.say(v.bool("enable"))
```

- web.context.say(string...)
```lua
    local ctx = web.context
    local r = web.router()

    r.GET("/" , function()
        ctx.say("helo")
    end)
```
- web.context.append(string...) 
```lua
    local ctx = web.context
    local r = web.router()
    r.GET("/" , function()
        ctx.say("helo")
        ctx.append(" vela-go")
    end)
```
- web.context.exit(int) 
```lua
    local ctx = web.context
    local r = web.router()
    r.GET("/" , function()
        ctx.say("helo")
        ctx.append(" vela-go")
        ctx.exit(200)
    end)
```
- web.context.set_header(name1 , value1 , name2, value2,name3 , value3)
```lua
  local ctx = web.context
  local r = web.router()
  r.GET("/" , function()
      ctx.set_healder("content-length",1 , "uuid","x-x")
      ctx.say("helo")
  end)
```
- web.context.eof()
```lua
    local ctx = web.context
    local r = web.router()

    r.GET("/" , function()
        ctx.say("helo")
        ctx.eof() --关闭继续运行
    end)
```

- web.context.arg_*
- 作用: 读取用户请求参数的值
- 语法： web.context.arg_name , web.context.arg_a
```lua
    -- http://www.a.com/a?name=edx&a=123
    local ctx = web.context
    local name = ctx.arg_name or "" --name=edx
    local a = ctx.arg_a or "" -- a=123
```

- web.context.post_*
- 作用： 读取用户POST参数
- 语法： web.context.post_value
```lua
    -- POST /api HTTP/1.1
    -- Host: www.a.com
    -- Rule: sqli-deny
    --
    -- name=edunx&value=123
    
    local ctx = web.context
    local say = http.response.say

    say(ctx.post_name , " " , ctx.post_value)
```

- web.context.param_*
- 作用： 读取路由中的param
- 语法： web.context.param_name , web.context.param_val
```lua
 -- http://www.a.com/api/admin/123456
 -- 路由 r.GET("/api/{name}/{val:*}

    local ctx = web.context

    local name = ctx.param_name 
    local val = ctx.param_val

    ctx.say(name , " " , val)
```

- web.context.http_*
- 作用：获取请求header头里面的参数
- 语法： web.context.http_rule , web.context.http_user_agent
```lua
    -- GET / HTTP/1.1
    -- Host: www.a.com
    -- Rule: sqli-deny

    local ctx = web.context
    local rule = ctx.http_rule
```
- web.context.cookie_*
- 作用： 获取cookie里的某个具体字段
- 语法： web.context.cookie_session

```lua
    -- GET / HTTP/1.1
    -- Host: www.a.com
    -- cookie:session=123x

    local ctx = web.context
    local v = ctx.cookie_session
    ctx.say(v)
```

- web.context.remote_addr
- 作用: 获取connection的四层IP地址
- 语法: web.context.remote_addr
```lua
    local addr = web.context.remote_addr
``` 
- web.context.remote_port
- 作用: 获取connection的四层端口
- 语法: web.context.remote_port
```lua
    local port = web.context.remote_port
``` 
- web.context.server_addr
- 作用: 获取本地服务器的IP地址
- 语法: web.context.server_addr
```lua
    local addr = web.context.server_addr
``` 
- web.context.server_port
- 作用: 获取本地服务器的端口
- 语法: web.context.server_addr
```lua
    local port = web.context.server_port
``` 
- web.context.host
- 作用: 获取用户请求的主机名
- 语法: web.context.host
```lua
    local host = web.context.host
``` 
- web.context.uri
- 作用: 获取获用户请求的URI
- 语法: web.context.uri
```lua
    -- http://www.a.com/api/info
    local uri = web.context.uri -- /api/info
``` 
- web.context.args
- 作用: 获取用户请求的args字符串
- 语法: web.context.args
```lua
    -- http://www.a.com/api/info?name=admin&val=123
    local args = web.context.args -- name=admin&val=123
``` 
- web.context.request_uri
- 作用: 获取完整的请求URI
- 语法: web.context.request_uri
```lua
    -- http://www.a.com/api/info?name=admin&val=123
    local request = web.context.request_uri -- /api/info?name=admin&val=123
``` 

- web.context.http_time
- 作用: 获取请求时间
- 语法: web.context.http_time
```lua
    local ht = web.context.http_time --2020-01-01 01:02:03.00
``` 
- web.context.cookie_raw
- 作用: 获取cooie的子完整自字符串
- 语法: web.context.cookie_raw
```lua
    local cookie = web.context.cookie_raw
``` 

- web.context.header_raw
- 作用: 获取完整的header字符串
- 语法: web.context.header_raw
```lua
    local raw = web.context.header_raw
``` 
- web.context.content_length
- 作用: 获取获取用户请求的包大小
- 语法: web.context.content_length
```lua
    local len = web.context.content_length
``` 
- web.context.content_type
- 作用: 获取获取用户请求的content_type
- 语法: web.context.content_type
```lua
    local ct = web.context.content_type
``` 
- web.context.body
- 作用: 获取用户的请求的body请求体
- 语法: web.context.body
```lua
    local body = web.context.body
``` 
- web.context.region_cityid
- 作用: 获取用户所在城市的ID
- 语法: web.context.region_cityid
```lua
    local id = web.context.region_cityid
``` 
- web.context.region_info
- 作用: 获取获IP地址位置信息
- 语法: web.context.region_info
```lua
    local info = web.context.region_info
``` 
- web.context.ua
- 作用: 获取user_agent
- 语法: web.context.ua
```lua
    local ua = web.context.ua
``` 
- web.context.referer
- 作用: 获取referer
- 语法: web.context.referer
```lua
    local ref = web.context.referer
``` 
- web.context.status
- 作用: 获取返回状态码
- 语法: web.context.status
```lua
    local status = web.context.status
``` 
- web.context.sent
- 作用: 获取发送数据包的大小
- 语法: web.context.sent
```lua
    local sent = web.context.sent
```

#### 3.handle配置
主要的业务处理逻辑 绑定之前注册路由
```lua
    local ctx = web.context
    web.handle{
        -- 新建过滤条件
        filter = web.filter{ 
           "$host == www.a.com,www.b.com", 
           "$uri == /123"
        },
    
        -- 返回状态码
        code = 200, 
        header = {
            ["content-length"] = 100,
            ["server"] = "hacker"
        },
    
        -- 返回值 
        -- 中间可以带有动态内容如：${uri} ${param_name}
        body = "hello ${uri} ${param_name}",
        
        -- body 可以直接是function
        body = function()
            
        end,

        -- 关闭 路由些其他handle的处理
        eof = true, 
    }
```
#### 4.handle 说明
handle 下面调用方式
```lua
    -- 注册路由的时候直接创建
    r.GET("/123" , web.handle{ })
    
    -- 注册路由时候 告诉名称， 自动从公共库中加载
    r.GET("/get/info" , "sqli" , "xss" , "not_found")
    
    --直接注入匿名函数
    r.GET("/get/useinfo" , function() end)

    r.GET("/get/useinfo" , web.handle'${app}')

    --文件路径
    r.FILE("/static/{filepath:*}" , "share/www/html")

    --文件重定向
    r.FILE("/static/{filepath:*}" , "share/www/html" , function()
        return "/index.html" 
    end)
```

#### 5.handle 库定义
handle库的定义，一般是文件名调用 查找路径 setup 中的 handler
例如下图中的handle
```lua
    return web.handle{ --记得一定要return 不然无法获取
        code = 200,
        header = web.header{},
        body = "aaaa${uri}"
    }
```