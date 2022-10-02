# Gee

原文链接：[7天用Go从零实现Web框架Gee教程 | 极客兔兔 (geektutu.com)](https://geektutu.com/post/gee.html)

本文为学习笔记

如何实现一个 Web 框架？

**Go 语言内置的 net/http 库，封装了HTTP网络编程的基础的接口，Gee Web 框架就是基于 net/http 的。**

在设计一个框架之前，我们需要回答框架核心为我们解决了什么问题。只有理解了这一点，才能想明白我们需要在框架中实现什么功能。

为什么要用框架？使用基础库时，需要频繁手工处理的地方，就是框架的价值所在。

Gee框架中很多部分实现的功能都很简单，但是尽可能地体现了一个框架核心的设计原则。

net/http提供了基础的Web功能，即监听端口，映射静态路由，解析HTTP报文。一些Web开发中简单的需求并不支持，需要手工实现。

动态路由：例如hello/:name，hello/*这类的规则。
鉴权：没有分组/统一鉴权的能力，需要在每个路由映射的handler中实现。
模板：没有统一简化的HTML机制。

Gee框架是用 Go语言实现的一个简单的 Web 教程，起名叫 Gee，为 geektutu.com 的前三个字母。

这个框架中的很多部分实现的功能都很简单，但是尽可能地体现一个框架核心的设计原则。



## day1 http.Handler

- 简单介绍`net/http`库以及`http.Handler`接口。
- 搭建`Gee`框架的雏形

http.HandleFunc() 用来注册路由，将处理函数绑定到指定路由

```go
// HandleFunc registers the handler function for the given pattern
// in the DefaultServeMux.
// The documentation for ServeMux explains how patterns are matched.
func HandleFunc(pattern string, handler func(ResponseWriter, *Request)) {
   DefaultServeMux.HandleFunc(pattern, handler)
}
```

返回的函数类型的参数中，ResponseWriter 是一个接口类型，HTTP handler 使用 ResponseWriter 接口构造 HTTP 响应。

```go
type ResponseWriter interface {
   Header() Header
   Write([]byte) (int, error)
   WriteHeader(statusCode int)
}
```

参数 Request，包含了该 HTTP 请求的所有信息

```Go
log.Fatal(http.ListenAndServe(":9999", nil))
```

这一行用来启动Web服务，第一个参数是地址，:9999表示在 9999 端口监听。第二个参数则代表处理所有 HTTP 请求的实例，nil 代表使用标准库中的实例处理。第二个参数，则是我们基于 net/http 标准库实现 Web 框架的入口。

函数 ListenAndServe 第二个参数是一个 Handler 接口，需要实现其方法 ServeHTTP，也就是说，只要传入任何实现了 ServeHTTP 接口的实例，所有的 HTTP 请求，就都交给该实例处理了。

接下来我们就声明一个实现 ServeHTTP 的实例，将所有请求交给该实例，将其命名为 Engine。

**Engine 是所有请求的统一处理程序**

定义为一个空的结构体 Engine，实现了方法 ServeHTTP，该方法有两个参数，（与 HandlerFunc 第二个参数类似）

- 第一个参数是 ResponseWriter，利用 ResponseWriter 可以构造针对该请求的响应
- 第二个参数是 Request，该对象包含了该 HTTP 请求的所有信息，比如请求地址，Header 和 Body 等信息

在main 函数中，我们给 ListenAndServe 方法的第二个参数传入了刚才创建的 engine 实例。至此，我们就走出了实现 Web 框架的第一步，即，将所有的 HTTP 请求转向了我们自己的处理逻辑，也就能针对具体的路由写处理程序。在实现 engine 后，我们拦截了所有的 HTTP 请求，拥有了统一的控制入口，在这里我们可以自由定义路由映射的规则，也可以统一添加一些处理逻辑，例如 日志、异常处理等。

使用 New() 创建 gee 的实例，GET() 或POST() 方法添加路由，最后使用 Run() 启动 Web 服务。这里的路由，还只是静态路由，不支持 /hello/:name 这样的动态路由，动态路由之后会实现。

那么 gee.go 就是重头戏了，重点介绍这部分的实现：

- 首先定义了类型HandlerFunc，这是提供给框架用户的，用来定义路由映射的处理方法。我们在Engine中，添加了一张路由映射表router，key 由请求方法和静态路由地址构成，例如GET-/、GET-/hello、POST-/hello，这样针对相同的路由，如果请求方法不同,可以映射不同的处理方法(Handler)，value 是用户映射的处理方法。
- 当用户调用 (Engine).GET() 方法时，会将路由和处理方法注册到映射表 router 中，(*Engine).Run()方法，是 ListenAndServe 的包装。
- Engine 实现的 ServeHTTP 方法的作用就是，解析请求的路径，查找路由映射表，如果查到，就执行注册的处理方法。如果查不到，就返回 404 NOT FOUND 。

至此，整个 Gee 框架的原型已经出来了。实现了路由映射表，提供了用户注册路由的方法，包装了启动服务的函数。

之后会慢慢将 动态路由、中间件 等功能添加上去。

在 Go 语言中，实现了接口方法的 struct 都可以强制转换为接口类型。

```Go
handler := (http.Handler)(engine) // 手动转换为借口类型
log.Fatal(http.ListenAndServe(":9999", handler))
```



## day2-上下文

- 将`路由(router)`独立出来，方便之后增强。
- 设计`上下文(Context)`，封装 Request 和 Response ，提供对 JSON、HTML 等返回类型的支持。
- 动手写 Gee 框架的第二天，**框架代码140行，新增代码约90行**

设计 Context 的必要性：

1. 对Web服务来说，无非是根据请求*http.Request，构造响应http.ResponseWriter。但是这两个对象提供的接口粒度太细，比如我们要构造一个完整的响应，需要考虑消息头(Header)和消息体(Body)，而 Header 包含了状态码(StatusCode)，消息类型(ContentType)等几乎每次请求都需要设置的信息。因此，如果不进行有效的封装，那么框架的用户将需要写大量重复，繁杂的代码，而且容易出错。针对常用场景，能够高效地构造出 HTTP 响应是一个好的框架必须考虑的点。
2. 针对使用场景，封装*http.Request和http.ResponseWriter的方法，简化相关接口的调用，只是设计 Context 的原因之一。对于框架来说，还需要支撑额外的功能。例如，将来解析动态路由/hello/:name，参数:name的值放在哪呢？再比如，框架需要支持中间件，那中间件产生的信息放在哪呢？Context 随着每一个请求的出现而产生，请求的结束而销毁，和当前请求强相关的信息都应由 Context 承载。因此，设计 Context 结构，扩展性和复杂性留在了内部，而对外简化了接口。路由的处理函数，以及将要实现的中间件，参数都统一使用 Context 实例， Context 就像一次会话的百宝箱，可以找到任何东西。

此外：

- `Handler`的参数变成成了`gee.Context`，提供了查询Query/PostForm参数的功能。
- `gee.Context`封装了`HTML/String/JSON`函数，能够快速构造HTTP响应。

具体实现，在 Context.go中

- 代码最开头，给map[string]interface{}起了一个别名gee.H，构建JSON数据时，显得更简洁。
- Context目前只包含了http.ResponseWriter和*http.Request，另外提供了对 Method 和 Path 这两个常用属性的直接访问。
- 提供了访问Query和PostForm参数的方法。
- 提供了快速构造String/Data/JSON/HTML响应的方法。

H 是一个自定义类型，引入 H 可以简化生成 json 的方式，如果需要嵌套 json，那么嵌套 gin.H 就可以了

路由 Router.go

我们将和路由相关的方法和结构提取了出来，放到了一个新的文件中`router.go`，方便我们下一次对 router 的功能进行增强，例如提供动态路由的支持。 router 的 handle 方法作了一个细微的调整，即 handler 的参数，变成了 Context。

将router相关的代码独立后，gee.go简单了不少。最重要的还是通过实现了 ServeHTTP 接口，接管了所有的 HTTP 请求。相比第一天的代码，这个方法也有细微的调整，在调用 router.handle 之前，构造了一个 Context 对象。这个对象目前还非常简单，仅仅是包装了原来的两个参数，之后我们会慢慢地给Context插上翅膀。



## day3-前缀树路由

- 使用 Trie 树实现动态路由(dynamic route)解析。
- 支持两种模式`:name`和`*filepath`，**代码约150行**。

Trie 树简介

之前，我们用了一个非常简单的 map 结构存储了路由表，使用map存储键值对，索引非常高效，但是有一个弊端，键值对的存储的方式，只能用来索引静态路由。那如果我们想支持类似于/hello/:name这样的动态路由怎么办呢？所谓动态路由，即一条路由规则可以匹配某一类型而非某一条固定的路由。例如/hello/:name，可以匹配/hello/geektutu、hello/jack等。

动态路由有很多种实现方式，支持的规则、性能等有很大的差异。例如开源的路由实现gorouter支持在路由规则中嵌入正则表达式，例如/p/[0-9A-Za-z]+，即路径中的参数仅匹配数字和字母；另一个开源实现httprouter就不支持正则表达式。著名的Web开源框架gin 在早期的版本，并没有实现自己的路由，而是直接使用了httprouter，后来不知道什么原因，放弃了httprouter，自己实现了一个版本。

实现动态路由最常用的数据结构，被称为前缀树(Trie树)。每一个节点的所有子节点都拥有相同的前缀，这种结构非常适用于路由匹配。

HTTP请求的路径恰好是由/分隔的多段构成的，因此，每一段可以作为前缀树的一个节点。我们通过树结构查询，如果中间某一层的节点都不满足条件，那么就说明没有匹配到的路由，查询结束。

接下来我们实现的动态路由具备以下两个功能。

- 参数匹配`:`。例如 `/p/:lang/doc`，可以匹配 `/p/c/doc` 和 `/p/go/doc`。
- 通配`*`。例如 `/static/*filepath`，可以匹配`/static/fav.ico`，也可以匹配`/static/js/jQuery.js`，这种模式常用于静态服务器，能够递归地匹配子路径。

### Trie 树实现

首先我们需要设计树节点上应该存储那些信息。

```go
type node struct {
   pattern  string       //待匹配的完整路由
   part     string       //路由的一部分
   children []*node      //子节点
   isWild   bool         //是否模糊匹配，实现动态路由的关键
}
```

与普通的树不同，为了实现动态路由匹配，加上了 isWild 这个参数。即当我们匹配 /p/go/doc/这个路由时，第一层节点，p精准匹配到了p，第二层节点，go模糊匹配到:lang，那么将会把lang这个参数赋值为go，继续下一层匹配。我们将匹配的逻辑，包装为一个辅助函数。

对于路由来说，最重要的当然是注册与匹配了。**开发服务时，注册路由规则，映射handler；访问时，匹配路由规则，查找到对应的handler。**因此，Trie 树需要支持节点的插入与查询。插入功能很简单，递归查找每一层的节点，如果没有匹配到当前part的节点，则新建一个。

在一个完整路由中，最后一部分的路由 pattern 值为 完整路由，其余部分的 pattern 值为空。当匹配结束后， 我们可以使用 n.pattern=="" 来判断路由规则是否匹配成功。

查询功能，同样也是递归查询每一层的节点，退出规则是，匹配到了*，匹配失败，或者匹配到了第len(parts)层节点。

### Router

Trie 树的插入与查找（对应注册和匹配路由）都成功实现了，接下来我们将 Trie 树应用到路由中去。我们使用 roots 来存储每种请求方式的Trie 树根节点。使用 handlers 存储每种请求方式的 HandlerFunc 。

getRoute 函数中，还解析了:和两种匹配符的参数，返回一个 map 。例如 /p/go/doc 匹配到 /p/:lang/doc，解析结果为：{lang: "go"}，/static/css/geektutu.css 匹配到 /static/filepath，解析结果为 {filepath: "css/geektutu.css"}。

### Context 与 handle 的变化

在 HandlerFunc 中，希望能够访问到解析的参数，因此，需要对 Context 对象增加一个属性和方法，来提供对路由参数的访问。我们将解析后的参数存储到Params中，通过c.Param("lang")的方式获取到对应的值。

router.go的变化比较小，比较重要的一点是，在调用匹配到的handler前，将解析出来的路由参数赋值给了c.Params。这样就能够在handler中，通过Context对象访问到具体的值了。



## day4-分组控制

**实现路由分组控制（Router Group Control），代码约 50 行**

### 分组的意义

分组控制(Group Control)是 Web 框架应提供的基础功能之一。所谓分组，是指路由的分组。如果没有路由分组，我们需要针对每一个路由进行控制。但是真实的业务场景中，往往某一组路由需要相似的处理。例如：

- 以 /post 开头的路由匿名可访问。
- 以 /admin 开头的路由需要鉴权。
- 以 /api 开头的路由是 RESTful 接口，可以对接第三方平台，需要三方平台鉴权。

大部分情况下的路由分组，是以相同的前缀来区分的。因此，我们今天实现的分组控制也是以前缀来区分，并且支持分组的嵌套。例如 /post 是一个分组，/post/a 和 /post/b 可以是该分组下的子分组。作用在 /post 分组上的中间件(middleware)，也都会作用在子分组，子分组还可以应用自己特有的中间件。

中间件可以给框架提供无限的扩展能力，应用在分组上，可以使得分组控制的收益更为明显，而不是共享相同的路由前缀这么简单。例如`/admin`的分组，可以应用鉴权中间件；`/`分组应用日志中间件，`/`是默认的最顶层的分组，也就意味着给所有的路由，即整个框架增加了记录日志的能力。

### 分组嵌套

一个 Group 对象需要具备哪些属性呢？首先是前缀(prefix)，比如`/`，或者`/api`；要支持分组嵌套，那么需要知道当前分组的父亲(parent)是谁；当然了，按照我们一开始的分析，中间件是应用在分组上的，那还需要存储应用在该分组上的中间件(middlewares)。还记得，我们之前调用函数`(*Engine).addRoute()`来映射所有的路由规则和 Handler 。如果Group对象需要直接映射路由规则的话，比如我们想在使用框架时，这么调用：

```go
r := gee.New()
v1 := r.Group("/v1")
v1.GET("/", func(c *gee.Context) {
	c.HTML(http.StatusOK, "<h1>Hello Gee</h1>")
})
```

那 Group 对象需要有访问 Router 的能力，因为它在这里直接注册路由了，怎么给它这个能力呢？我们可以给在 Group 结构体中，保存一个指向 Engine 的指针，借用 Engine 来注册路由。

整个框架的资源都是由 Engine 统一协调的，那么就可以通过 Engine 间接地访问各种接口了。

最终，Group 的定义如下：

```go
RouterGroup struct {
   prefix      string
   middlewares []HandlerFunc // support middleware
   parent      *RouterGroup  // support nesting 
   engine      *Engine       // all groups share a Engine instance
}
```

我们还可以进一步抽象，将 Engine 作为最顶层的分组，也就是说 Engine 还要拥有 RouterGroup 的所有能力

```go
Engine struct {
   *RouterGroup
   router *router
   groups []*RouterGroup // store all groups
}
```

接下来我们就可以将和路由有关的函数，都交给 RouterGroup 实现，而不是由 Engine 来操作。

此外，我们创建新的路由分组 还是通过 Engine 创建的，是 Engine 的方法。

可以仔细观察下addRoute函数，调用了group.engine.router.addRoute来实现了路由的映射。由于Engine从某种意义上继承了RouterGroup的所有属性和方法，因为 (*Engine).engine 是指向自己的。这样实现，我们既可以像原来一样添加路由，也可以通过分组添加路由。



## day5-中间件

- 设计并实现 Web 框架的中间件（Middlewares）机制
- 实现通用的 Logger 中间件，能够记录请求到响应所花费的时间

### 中间件是什么？

中间件，简单来说，就是非业务的技术类组件。Web 框架本身不可能理解所有的业务，因而不可能实现所有的功能。因此，框架需要有一个插口，允许用户自己定义功能，嵌入到框架中，仿佛这个功能是框架原生支持的一样。因此，对于中间件而言，需要考虑 2 个比较关键的点：

- 插入点在哪？使用框架的人并不关心底层逻辑的具体实现，如果插入点太底层，中间逻辑就会非常复杂。如果插入点离用户太近，那和用户直接定义一组函数，每次在 Handler 中手工调用没多大区别。
- 中间件的输入是什么？中间件的输入，决定了扩展能力。暴露的参数太少，用户发挥空间有限。

Gee 框架中间件的设计参考了 Gin 框架。

### 中间件设计

Gee 的中间件的定义与路由映射的 Handler 一致，，处理的输入是 Context 对象，插入点是 Gee 初始化 Context 对象后，允许用户使用自己定义的中间件做一些额外的处理，例如记录日志、统计处理请求时长，对 Context 二次加工等。

通过调用 （*Context）.Next() 函数，中间件可等待用户自定义的 Handler 处理结束后，再做一些额外的操作，比如计算本次处理所有时长等。即 Gee 的中间件支持用户在请求被处理的前后，在做一些额外的操作。

举个例子，我们希望最终能够支持如下定义的中间件，`c.Next()`表示等待执行其他的中间件或用户的`Handler`：

```go
func Logger() HandlerFunc {
	return func(c *Context) {
		// Start timer
		t := time.Now()
		// Process request
		c.Next()
		// Calculate resolution time
		log.Printf("[%d] %s in %v", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}
```

另外，支持设置多个中间件，依次进行调用。

中间件是应用在 RouterGroup 上的，应用在最顶层的 Group，相当于作用于全局，所有请求都会被该中间件处理。那为什么不作用在每一条路由规则上呢？作用在某条路由规则，那还不如用户直接在 Handler 中调用直观。只作用在某条路由规则的功能通用性太差，不适合定义为中间件。

我们之前的框架设计是这样的，当接收到请求后，匹配路由，该请求的所有信息都保存在Context中。中间件也不例外，接收到请求后，应查找所有应作用于该路由的中间件，保存在Context中，依次进行调用。为什么依次调用后，还需要在Context中保存呢？因为在设计中，中间件不仅作用在处理流程前，也可以作用在处理流程后，即在用户定义的 Handler 处理完毕后，还可以执行剩下的操作。

为此，我们给Context添加了2个参数（handlers 和 index），定义了Next方法。

```go
func (c *Context) Next() {
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}
```

**index是记录当前执行到第几个中间件，当在中间件中调用Next方法时，控制权交给了下一个中间件，直到调用到最后一个中间件，然后再从后往前，调用每个中间件在Next方法之后定义的部分。**

最后，定义 Use 函数，将中间件应用到某个 RouterGroup 。

ServeHTTP 函数也有变化，当我们接收到一个具体请求时，要判断该请求适用于哪些中间件，在这里我们简单通过 URL 的前缀来判断。得到中间件列表后，赋值给 `c.handlers`。

handle 函数中，将从路由匹配得到的 Handler 添加到 `c.handlers`列表中，执行`c.Next()`。



## day6-模板（HTML Template）

- 实现静态资源服务(Static Resource)。
- 支持HTML模板渲染。

浅析 服务端渲染/前后端渲染/客户端渲染：https://blog.csdn.net/weixin_42207975/article/details/106967494 

现在非常流行前后端分离的开发模式，即 Web 后端提供 Restful 接口，返回结构化的数据（通常为 JSON 或 XML），前端使用 AJAX 技术请求到所需的数据，利用 JavaScript 进行渲染。Vue/React 等前端框架持续火热，这种开发模式前后端解耦，优势非常突出。

后端开发人员专心解决资源利用、并发、数据库等问题，只需要考虑数据如何生成；

前端开发人员专注于界面设计实现，只需要考虑拿到数据后如何渲染即可。

而且，前后端分离在当前还有一个不可忽视的优势。因为后端只关注于数据，接口返回值是结构化的，与前端解耦。同一套后端服务能够同时支撑小程序、移动APP、PC端Web页面，以及对外提供的接口。

随着前端工程化的不断发展，前端技术越来越自成体系了。

前后端分离的一个问题在于，页面是在客户端渲染的，比如浏览器，这对爬虫来说并不友好。

今天的内容就是介绍 Web 狂阿基如何支持服务端渲染的场景。

### 静态文件

网页的三剑客，JavaScript、CSS 和 HTML。要做到服务端渲染，第一步便是要支持 JS、CSS 等静态文件。还记得我们之前设计动态路由的时候，支持通配符`*`匹配多级子路径。比如路由规则`/assets/*filepath`，可以匹配`/assets/`开头的所有的地址。例如`/assets/js/geektutu.js`，匹配后，参数`filepath`就赋值为`js/geektutu.js`。

那如果我么将所有的静态文件放在`/usr/web`目录下，那么`filepath`的值即是该目录下文件的相对地址。映射到真实的文件后，将文件返回，静态服务器就实现了。

找到文件后，如何返回这一步，`net/http`库已经实现了。因此，gee 框架要做的，仅仅是解析请求的地址，映射到服务器上文件的真实地址，交给`http.FileServer`处理就好了。

我们给`RouterGroup`添加了2个方法，createStaticHandler 和 Static，`Static`这个方法是暴露给用户的。用户可以将磁盘上的某个文件夹`root`映射到路由`relativePath`。例如：

```go
r := gee.New()
r.Static("/assets", "/usr/geektutu/blog/static")
// 或相对路径 r.Static("/assets", "./static")
r.Run(":9999")
```

用户访问`localhost:9999/assets/js/geektutu.js`，最终返回`/usr/geektutu/blog/static/js/geektutu.js`。

### HTML模板渲染

Go语言内置了`text/template`和`html/template`2个模板标准库，其中[html/template](https://golang.org/pkg/html/template/)为 HTML 提供了较为完整的支持。包括普通变量渲染、列表渲染、对象渲染等。gee 框架的模板渲染直接使用了`html/template`提供的能力。

```go
Engine struct {
	*RouterGroup
	router        *router
	groups        []*RouterGroup     // store all groups
	htmlTemplates *template.Template // for html render
	funcMap       template.FuncMap   // for html render
}

func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}
```

首先为 Engine 示例添加了 `*template.Template` 和 `template.FuncMap`对象，前者将所有的模板加载进内存，后者是所有的自定义模板渲染函数。

另外，给用户分别提供了设置自定义渲染函数`funcMap`和加载模板的方法。

接下来，对原来的 `(*Context).HTML()`方法做了些小修改，使之支持根据模板文件名选择模板进行渲染。

```go
type Context struct {
    // ...
	// engine pointer
	engine *Engine
}

func (c *Context) HTML(code int, name string, data interface{}) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(500, err.Error())
	}
}
```

我们在 `Context` 中添加了成员变量 `engine *Engine`，这样就能够通过 Context 访问 Engine 中的 HTML 模板。实例化 Context 时，还需要给 `c.engine` 赋值。

```go
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// ...
	c := newContext(w, req)
	c.handlers = middlewares
	c.engine = engine
	engine.router.handle(c)
}
```



## day7-错误恢复（Panic Recover）

### panic、recover 和 recover

Go 语言中，比较常见的错误处理方法是返回 error，由调用者决定后续如何处理。但是如果是无法恢复的错误，可以手动触发 panic，当然如果在程序运行过程中出现了类似于数组越界的错误，panic 也会被触发。

panic 会导致程序被中止，但是在退出前，会先处理完当前协程上已经defer 的任务，执行完成后再退出。

可以 defer 多个任务，在同一个函数中 defer 多个任务，会逆序执行。即先执行最后 defer 的任务。

Go 语言还提供了 recover 函数，可以避免因为 panic 发生而导致整个程序终止，recover 函数只在 defer 中生效。

### Gee 的错误处理机制

对一个 Web 框架而言，错误处理机制是非常必要的。可能是框架本身没有完备的测试，导致在某些情况下出现空指针异常等情况，也有可能用户不正确的参数，出发了某些异常，例如数组越界，空指针等。如果因为这些原因导致系统宕机，必然是不可接受的。

我们将在 Gee 中添加一个非常简单的错误处理机制，即在错误发生时，向用户返回 Internal Server Error，并且在日志中打印必要的错误信息，方便进行错误定位。

我们之前实现了中间件机制，错误处理也可以作为一个中间件，增强 gee 框架的能力。

`Recovery` 的实现非常简单，使用 defer 挂载上错误恢复的函数，在这个函数中调用 *recover()*，捕获 panic，并且将堆栈信息打印在日志中，向用户返回 *Internal Server Error*。

trace() 函数，用于获取触发 panic 的堆栈信息

```go
// print stack trace for debug
func trace(message string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:]) // skip first 3 caller

	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}

func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				log.Printf("%s\n\n", trace(message))
				c.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()

		c.Next()
	}
}
```
