# PLAN: opencode web ctx base 支持

## 问题

opencode web 构建时 `base="/"`, 所有路径（assets、API、路由）都基于根路径 `/`。
通过 piko 在子路径 `/developer-tools/` 下代理时，导致大量 404。

之前的方案（proxy 层面 body 替换 + runtime interceptor）太 hack，容易漏。

## 新方案：修改 opencode 源码原生支持 base path

### 1. 添加 opencode 为 git submodule

```bash
git submodule add git@github.com:friddle/opencode.git
```

### 2. 修改 opencode web（3 处）

#### 2a. `packages/app/vite.config.ts` — 支持 build-time base

```ts
export default defineConfig({
  base: process.env.OPENCODE_WEB_BASE ?? "/",  // 新增
  // ...
})
```

Build 时传 `OPENCODE_WEB_BASE=/developer-tools`，Vite 会：
- HTML 里所有 `<script src>`, `<link href>` 自动加前缀
- 动态 import 的 `__vite__mapDeps` 路径自动加前缀
- CSS `url()` 引用自动加前缀

#### 2b. `packages/app/src/entry.tsx:102-107` — server URL 加 base

```ts
const getCurrentUrl = () => {
  if (location.hostname.includes("opencode.ai")) return "http://localhost:4096"
  if (import.meta.env.DEV)
    return `http://...`
  return location.origin + import.meta.env.BASE_URL   // 原来是 location.origin
}
```

这样 API 调用 (`/session`, `/agent`, `/global/health`) 的 baseUrl 就包含了 `/developer-tools` 前缀。

#### 2c. `packages/app/src/app.tsx:443-444` — Router 加 base

```tsx
<Dynamic
  component={props.router ?? Router}
  base={import.meta.env.BASE_URL.replace(/\/$/, "")}  // 新增
  // ...
>
```

SolidJS Router 的 `base` prop 会：
- 从 `location.pathname` 中剥离 base 前缀来匹配路由
- `navigate()` / `<A href>` 自动加前缀

### 3. 构建流程

#### opencode-piko-remote/client/Makefile

`download-opencode` 改为从 submodule 构建：

```makefile
OPENCODE_WEB_BASE ?= /

build-opencode-web:
    cd ../opencode/packages/app
    OPENCODE_WEB_BASE=$(OPENCODE_WEB_BASE) bun run build
    # 产物在 dist/，打包进 opencode binary

build-opencode-binary:
    cd ../opencode
    # 把 web dist embed 到 Go binary
    go build -o opencode ./packages/opencode
```

运行时流程：
1. opencode-piko 启动，知道 endpoint name = `developer-tools`
2. 构建/解压 opencode binary（base 已 baked 为 `/developer-tools`）
3. 启动 opencode web
4. rewrite proxy 只做 path stripping

### 4. 简化 middleware.go

删掉所有 body rewriting 和 interceptor，只保留 path stripping：

```go
proxy := &httputil.ReverseProxy{
    Rewrite: func(pr *httputil.ProxyRequest) {
        pr.SetURL(target)
        pr.Out.Host = pr.In.Host
        // 唯一需要做的事：strip endpoint prefix
        if strings.HasPrefix(pr.Out.URL.Path, prefix+"/") {
            pr.Out.URL.Path = strings.TrimPrefix(pr.Out.URL.Path, prefix)
        } else if pr.Out.URL.Path == prefix {
            pr.Out.URL.Path = "/"
        }
    },
    // 删掉 ModifyResponse — 不再需要
}
```

删掉：
- `rewriteHTML()` / `rewriteJS()`
- `injectScriptRef()` / `interceptorScript()`
- `Handler()` 里的 interceptor file serving
- CSP header removal

### 5. 每个端点独立构建的问题

base path 在 Vite build 时 baked，但每个 opencode-piko 实例的 endpoint name 不同。

解决方案：**运行时注入 base，不走 Vite build-time base**

更灵活的方案：
1. Vite 不设 `base`（保持 `/`）
2. opencode web server 接受 `OPENCODE_WEB_BASE` 环境变量
3. server 在返回 HTML 时动态注入：
   ```html
   <base href="/developer-tools/">
   <script>window.__OPENCODE_BASE__="/developer-tools"</script>
   ```
4. app 读 `window.__OPENCODE_BASE__` 而不是 `import.meta.env.BASE_URL`
5. server 对 assets 请求做 path stripping（内部转发 `/developer-tools/assets/x.js` → `/assets/x.js`）

这样不需要 per-endpoint 构建。

## 实施步骤

1. [ ] `git submodule add git@github.com:friddle/opencode.git`
2. [ ] 修改 `packages/app/src/entry.tsx` — `getCurrentUrl()` 加 base
3. [ ] 修改 `packages/app/src/app.tsx` — Router 加 `base` prop
4. [ ] 修改 opencode server — 支持 `OPENCODE_WEB_BASE` 运行时注入
5. [ ] 修改 `client/Makefile` — 从 submodule 构建 opencode
6. [ ] 简化 `client/src/middleware.go` — 只保留 path stripping
7. [ ] 测试验证
