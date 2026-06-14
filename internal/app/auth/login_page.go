package auth

import (
	"bytes"
	"html/template"
)

type LoginPageOptions struct {
	Next      string
	ShowError bool
}

func LoginPageHTML(opts LoginPageOptions) ([]byte, error) {
	errorText := ""
	if opts.ShowError {
		errorText = "密钥无效"
	}
	data := struct {
		Next      string
		ErrorText string
	}{
		Next:      SafeRedirect(opts.Next),
		ErrorText: errorText,
	}
	var out bytes.Buffer
	if err := loginPageTemplate.Execute(&out, data); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

var loginPageTemplate = template.Must(template.New("runtime-login").Parse(`<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Proxy Runtime 登录</title>
  <style>
    :root { color-scheme: dark; font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; }
    body { margin: 0; min-height: 100vh; display: grid; place-items: center; background: radial-gradient(circle at top, #1f2a44 0, #0b1020 45%, #05070d 100%); color: #e5e7eb; }
    main { width: min(92vw, 420px); border: 1px solid rgba(148, 163, 184, .2); border-radius: 24px; padding: 28px; background: rgba(15, 23, 42, .82); box-shadow: 0 24px 80px rgba(0, 0, 0, .35); backdrop-filter: blur(18px); }
    h1 { margin: 0 0 8px; font-size: 24px; }
    p { margin: 0 0 22px; color: #94a3b8; }
    label { display: block; margin-bottom: 8px; color: #cbd5e1; font-size: 14px; }
    input { box-sizing: border-box; width: 100%; border: 1px solid rgba(148, 163, 184, .25); border-radius: 14px; padding: 13px 14px; background: rgba(2, 6, 23, .7); color: #f8fafc; outline: none; }
    input:focus { border-color: #38bdf8; box-shadow: 0 0 0 3px rgba(56, 189, 248, .18); }
    button { width: 100%; margin-top: 18px; border: 0; border-radius: 14px; padding: 13px 16px; color: #031018; background: linear-gradient(135deg, #38bdf8, #22c55e); font-weight: 700; cursor: pointer; }
    .error { margin-bottom: 14px; border: 1px solid rgba(248, 113, 113, .28); border-radius: 12px; padding: 10px 12px; color: #fecaca; background: rgba(127, 29, 29, .25); }
  </style>
</head>
<body>
  <main>
    <h1>Proxy Runtime</h1>
    <p>输入控制面密钥后继续访问 MetaCubeXD。</p>
    {{if .ErrorText}}<div class="error">{{.ErrorText}}</div>{{end}}
    <form method="post" action="/api/auth/login?next={{urlquery .Next}}">
      <input type="hidden" name="next" value="{{.Next}}">
      <label for="token">控制面密钥</label>
      <input id="token" name="token" type="password" autocomplete="current-password" autofocus required>
      <button type="submit">登录</button>
    </form>
  </main>
</body>
</html>`))
