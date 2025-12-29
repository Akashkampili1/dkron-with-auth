package dkron

import (
	"crypto/hmac"
	"crypto/sha256"
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const uiPathPrefix = "ui/"

//go:embed ui-dist
var uiDist embed.FS

// UI registers UI specific routes on the gin RouterGroup.
func (h *HTTPTransport) UI(r *gin.RouterGroup, aclEnabled bool) {
	// If we are visiting from a browser redirect to the dashboard
	r.GET("/", func(c *gin.Context) {
		switch c.NegotiateFormat(gin.MIMEHTML) {
		case gin.MIMEHTML:
			c.Redirect(http.StatusSeeOther, "/ui/")
		default:
			c.AbortWithStatus(http.StatusNotFound)
		}
	})

	if h.agent.config.UIAuthEnabled && h.agent.config.UISessionEnabled {
		r.GET("/login", func(c *gin.Context) {
			html := `<!doctype html><html><head><meta charset="utf-8"><title>Dkron Login</title></head><body><button id="b">Login</button><script>document.getElementById('b').addEventListener('click', async function(){const u=prompt('Username');if(!u)return;const p=prompt('Password');if(p===null)return;const r=await fetch('/ui/login',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({username:u,password:p})});if(r.ok){location.href='/ui/';}else{alert('Login failed');}});</script></body></html>`
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
		})
		r.POST("/ui/login", h.uiLoginHandler)
		r.POST("/ui/logout", h.uiLogoutHandler)
	}

	ui := r.Group("/" + uiPathPrefix)
	if h.agent.config.UIAuthEnabled && h.agent.config.UISessionEnabled {
		ui.Use(h.uiSessionAuthMiddleware())
	}

	assets, err := fs.Sub(uiDist, "ui-dist")
	if err != nil {
		h.logger.Fatal(err)
	}
	a, err := assets.Open("index.html")
	if err != nil {
		h.logger.Fatal(err)
	}
	b, err := io.ReadAll(a)
	if err != nil {
		h.logger.Fatal(err)
	}
	t, err := template.New("index.html").Parse(string(b))
	if err != nil {
		h.logger.Fatal(err)
	}
	h.Engine.SetHTMLTemplate(t)

	ui.GET("/*filepath", func(ctx *gin.Context) {
		p := ctx.Param("filepath")
		f := strings.TrimPrefix(p, "/")
		_, err := assets.Open(f)
		if err == nil && p != "/" && p != "/index.html" {
			ctx.FileFromFS(p, http.FS(assets))
		} else {
			jobs, err := h.agent.Store.GetJobs(ctx.Request.Context(), nil)
			if err != nil {
				h.logger.Error(err)
			}
			var (
				totalJobs                                   = len(jobs)
				successfulJobs, failedJobs, untriggeredJobs int
			)
			for _, j := range jobs {
				if j.Status == "success" {
					successfulJobs++
				} else if j.Status == "failed" {
					failedJobs++
				} else if j.Status == "" {
					untriggeredJobs++
				}
			}
			l, err := h.agent.leaderMember()
			ln := "no leader"
			if err != nil {
				h.logger.Error(err)
			} else {
				ln = l.Name
			}
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"DKRON_API_URL":            fmt.Sprintf("../%s", apiPathPrefix),
				"DKRON_LEADER":             ln,
				"DKRON_TOTAL_JOBS":         totalJobs,
				"DKRON_FAILED_JOBS":        failedJobs,
				"DKRON_UNTRIGGERED_JOBS":   untriggeredJobs,
				"DKRON_SUCCESSFUL_JOBS":    successfulJobs,
				"DKRON_ACL_ENABLED":        aclEnabled,
				"DKRON_UI_AUTH_ENABLED":    h.agent.config.UIAuthEnabled,
				"DKRON_UI_SESSION_ENABLED": h.agent.config.UISessionEnabled,
			})
		}
	})
}

func (h *HTTPTransport) uiSessionAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check session cookie first
		if cookie, err := c.Request.Cookie("dkron_ui_session"); err == nil {
			parts := strings.Split(cookie.Value, ":")
			if len(parts) == 3 {
				user := parts[0]
				expStr := parts[1]
				sig := parts[2]
				exp, err := strconv.ParseInt(expStr, 10, 64)
				if err == nil && time.Now().Unix() <= exp {
					expected := h.signUISession(user, exp)
					if sig == expected {
						c.Next()
						return
					}
				}
			}
		}

		// No valid session cookie: use sequential alert prompts for username and password
		target := c.Request.RequestURI
		html := `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1" /><title>Dkron Login</title><style>*,*:before,*:after{box-sizing:border-box}body{margin:0;font-family:system-ui,-apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;background:radial-gradient(1000px 500px at 20% -10%,#1a1c5a,#0b0c2e);color:#eaeaf2;min-height:100vh;display:flex;align-items:center;justify-content:center;position:relative;overflow:hidden}.grid{position:absolute;inset:0;background-image:linear-gradient(rgba(255,255,255,.06) 1px,transparent 1px),linear-gradient(90deg,rgba(255,255,255,.06) 1px,transparent 1px);background-size:24px 24px;opacity:.25;pointer-events:none}.card{width:360px;background:#15163f;border-radius:14px;box-shadow:0 18px 40px rgba(0,0,0,.35);padding:26px 24px}.brand{display:flex;align-items:center;justify-content:center;gap:8px;margin-bottom:10px}.logo{width:30px;height:30px;border-radius:50%;display:inline-flex;align-items:center;justify-content:center;background:#4c5bdc;color:#fff;font-weight:700}.title{margin:0;font-size:18px;color:#fff;text-align:center}.subtitle{margin:6px 0 20px;font-size:13px;color:#b7b8d6;text-align:center}.field{display:flex;flex-direction:column;margin-bottom:12px}.label{font-size:12px;color:#b7b8d6;margin-bottom:6px}input{width:100%;padding:10px 12px;border:1px solid #2a2b55;border-radius:10px;background:#0f1040;color:#fff;outline:none}input::placeholder{color:#8d8fb4}.row{display:flex;gap:8px;align-items:center;justify-content:space-between}.actions{display:flex;gap:8px;margin-top:6px}.btn{flex:1;padding:10px 12px;border:0;border-radius:10px;background:#4c5bdc;color:#fff;font-weight:600;cursor:pointer}.btn.secondary{background:#25275b}.btn:disabled{opacity:.6;cursor:not-allowed}.error{color:#ff6b6b;font-size:12px;margin-top:8px;text-align:center}</style></head><body><div class="grid"></div><div class="card"><div class="brand"><div class="logo">D</div><h1 class="title">Dkron</h1></div><p class="subtitle">Sign in to access the dashboard</p><form id="f" autocomplete="on"><div class="field"><label for="username" class="label">Username</label><input id="username" name="username" placeholder="Username" autocomplete="username" /></div><div class="field"><label for="password" class="label">Password</label><div class="row"><input id="password" name="password" type="password" placeholder="Password" autocomplete="current-password" /><button type="button" id="toggle" class="btn secondary" aria-label="Show password">Show</button></div></div><div class="actions"><button type="submit" id="btn" class="btn">Login</button></div><div class="error" id="err" style="display:none"></div></form></div><script>const f=document.getElementById('f'),b=document.getElementById('btn'),e=document.getElementById('err'),pw=document.getElementById('password'),tg=document.getElementById('toggle');tg.addEventListener('click',function(){if(pw.type==='password'){pw.type='text';tg.textContent='Hide';tg.setAttribute('aria-label','Hide password');}else{pw.type='password';tg.textContent='Show';tg.setAttribute('aria-label','Show password');}});f.addEventListener('submit',async function(ev){ev.preventDefault();e.style.display='none';b.disabled=true;const u=this.username.value.trim(),p=pw.value;try{const r=await fetch('/ui/login',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({username:u,password:p})});if(r.ok){location.href='` + template.HTMLEscapeString(target) + `';}else{e.textContent='Invalid credentials';e.style.display='block';} }catch(err){e.textContent='Login error';e.style.display='block';} finally{b.disabled=false;}});</script></body></html>`
		c.Data(http.StatusUnauthorized, "text/html; charset=utf-8", []byte(html))
		c.Abort()
	}
}

func (h *HTTPTransport) signUISession(user string, exp int64) string {
	m := fmt.Sprintf("%s:%d", user, exp)
	mac := hmac.New(sha256.New, []byte(h.agent.config.UISessionSecret))
	mac.Write([]byte(m))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func (h *HTTPTransport) uiLoginHandler(c *gin.Context) {
	type creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var cr creds
	if err := c.BindJSON(&cr); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if !h.agent.config.UISessionEnabled {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	if cr.Username != h.agent.config.UIAuthUsername || cr.Password != h.agent.config.UIAuthPassword {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	ttl := h.agent.config.UISessionTTL
	if ttl <= 0 {
		ttl = time.Hour
	}
	exp := time.Now().Add(ttl).Unix()
	sig := h.signUISession(cr.Username, exp)
	val := fmt.Sprintf("%s:%d:%s", cr.Username, exp, sig)
	http.SetCookie(c.Writer, &http.Cookie{Name: "dkron_ui_session", Value: val, Path: "/", MaxAge: int(ttl.Seconds()), HttpOnly: true})
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *HTTPTransport) uiLogoutHandler(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{Name: "dkron_ui_session", Value: "", Path: "/", MaxAge: -1, HttpOnly: true})
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
