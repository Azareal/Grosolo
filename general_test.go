package main

import (
	"bytes"
	"database/sql"
	"log"
	"net/http"
	"net/http/httptest"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"

	c "github.com/Azareal/Gosora/common"
	e "github.com/Azareal/Gosora/extend"
	"github.com/Azareal/Gosora/install"
	qgen "github.com/Azareal/Gosora/query_gen"
	"github.com/Azareal/Gosora/routes"
)

//var dbTest *sql.DB
var dbProd *sql.DB
var gloinited bool
var installAdapter install.InstallAdapter

func ResetTables() (err error) {
	err = installAdapter.InitDatabase()
	if err != nil {
		return errors.WithStack(err)
	}

	err = installAdapter.TableDefs()
	if err != nil {
		return errors.WithStack(err)
	}

	err = installAdapter.CreateAdmin()
	if err != nil {
		return errors.WithStack(err)
	}

	return installAdapter.InitialData()
}

func gloinit() (e error) {
	if gloinited {
		return nil
	}
	// TODO: Make these configurable via flags to the go test command
	c.Dev.DebugMode = false
	c.Dev.SuperDebug = false
	c.Dev.TemplateDebug = false
	qgen.LogPrepares = false
	//nogrouplog = true
	c.StartTime = time.Now()

	ws := func(e error) error {
		return errors.WithStack(e)
	}
	if e = c.LoadConfig(); e != nil {
		return ws(e)
	}
	if e = c.ProcessConfig(); e != nil {
		return ws(e)
	}
	c.Tasks = c.NewScheduledTasks()

	e = c.InitTemplates()
	if e != nil {
		return ws(e)
	}
	c.Themes, e = c.NewThemeList()
	if e != nil {
		return ws(e)
	}
	c.TopicListThaw = c.NewTestThaw()
	c.SwitchToTestDB()

	var ok bool
	installAdapter, ok = install.Lookup(dbAdapter)
	if !ok {
		return ws(errors.New("We couldn't find the adapter '" + dbAdapter + "'"))
	}
	installAdapter.SetConfig(c.DbConfig.Host, c.DbConfig.Username, c.DbConfig.Password, c.DbConfig.Dbname, c.DbConfig.Port)

	e = ResetTables()
	if e != nil {
		return e
	}
	e = InitDatabase()
	if e != nil {
		return e
	}
	e = afterDBInit()
	if e != nil {
		return e
	}

	rrcfg := rcfg()
	rrcfg.DisableTick = false
	router, e = NewGenRouter(rrcfg)
	if e != nil {
		return ws(e)
	}
	gloinited = true
	return nil
}

func rcfg() *RouterConfig {
	return &RouterConfig{
		Uploads:     http.FileServer(http.Dir("./uploads")),
		DisableTick: true,
	}
}

func init() {
	if e := gloinit(); e != nil {
		log.Print("Something bad happened")
		//debug.PrintStack()
		log.Fatalf("%+v\n", e)
	}
}

var benchTidI = 1
var benchTid = "1"

// TODO: Swap out LocalError for a panic for this?
func BenchmarkTopicAdminRouteParallel(b *testing.B) {
	binit(b)
	cfg := NewStashConfig()
	c.Dev.DebugMode = false
	c.Dev.SuperDebug = false

	admin, err := c.Users.Get(1)
	if err != nil {
		b.Fatal(err)
	}
	if !admin.IsAdmin {
		b.Fatal("UID1 is not an admin")
	}
	adminUIDCookie := http.Cookie{Name: "uid", Value: "1", Path: "/", MaxAge: c.Year}
	adminSessionCookie := http.Cookie{Name: "session", Value: admin.Session, Path: "/", MaxAge: c.Year}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			reqAdmin := httptest.NewRequest("get", "/topic/hm."+benchTid, bytes.NewReader(nil))
			reqAdmin.AddCookie(&adminUIDCookie)
			reqAdmin.AddCookie(&adminSessionCookie)

			// Deal with the session stuff, etc.
			user, ok := c.PreRoute(w, reqAdmin)
			if !ok {
				b.Fatal("Mysterious error!")
			}
			head, e := c.UserCheck(w, reqAdmin, &user)
			if e != nil {
				b.Fatal(e)
			}
			//w.Body.Reset()
			routes.ViewTopic(w, reqAdmin, &user, head, "1")
			if w.Code != 200 {
				b.Log(w.Body)
				b.Fatal("HTTP Error!")
			}
		}
	})

	cfg.Restore()
}

func BenchmarkTopicAdminRouteParallelWithRouter(b *testing.B) {
	binit(b)
	router, e := NewGenRouter(rcfg())
	if e != nil {
		b.Fatal(e)
	}
	cfg := NewStashConfig()
	c.Dev.DebugMode = false
	c.Dev.SuperDebug = false

	admin, e := c.Users.Get(1)
	if e != nil {
		b.Fatal(e)
	}
	if !admin.IsAdmin {
		b.Fatal("UID1 is not an admin")
	}
	uidCookie := http.Cookie{Name: "uid", Value: "1", Path: "/", MaxAge: c.Year}
	sessionCookie := http.Cookie{Name: "session", Value: admin.Session, Path: "/", MaxAge: c.Year}
	path := "/topic/hm." + benchTid

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			reqAdmin := httptest.NewRequest("get", path, bytes.NewReader(nil))
			reqAdmin.AddCookie(&uidCookie)
			reqAdmin.AddCookie(&sessionCookie)
			reqAdmin.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36")
			reqAdmin.Header.Set("Host", "localhost")
			reqAdmin.Host = "localhost"
			//w.Body.Reset()
			router.ServeHTTP(w, reqAdmin)
			if w.Code != 200 {
				b.Log(w.Body)
				b.Fatal("HTTP Error!")
			}
		}
	})

	cfg.Restore()
}

func BenchmarkTopicAdminRouteParallelAlt(b *testing.B) {
	BenchmarkTopicAdminRouteParallel(b)
}

func BenchmarkTopicAdminRouteParallelWithRouterAlt(b *testing.B) {
	BenchmarkTopicAdminRouteParallelWithRouter(b)
}

func BenchmarkTopicAdminRouteParallelAltAlt(b *testing.B) {
	BenchmarkTopicAdminRouteParallel(b)
}

func BenchmarkTopicGuestAdminRouteParallelWithRouterPre(b *testing.B) {
	runtime.GC()
}

func BenchmarkTopicGuestAdminRouteParallelWithRouter(b *testing.B) {
	binit(b)
	router, e := NewGenRouter(rcfg())
	if e != nil {
		b.Fatal(e)
	}
	cfg := NewStashConfig()
	c.Dev.DebugMode = false
	c.Dev.SuperDebug = false

	admin, e := c.Users.Get(1)
	if e != nil {
		b.Fatal(e)
	}
	if !admin.IsAdmin {
		b.Fatal("UID1 is not an admin")
	}
	uidCookie := http.Cookie{Name: "uid", Value: "1", Path: "/", MaxAge: c.Year}
	sessionCookie := http.Cookie{Name: "session", Value: admin.Session, Path: "/", MaxAge: c.Year}
	path := "/topic/hm." + benchTid
	//runtime.GC()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			reqAdmin := httptest.NewRequest("get", path, bytes.NewReader(nil))
			reqAdmin.AddCookie(&uidCookie)
			reqAdmin.AddCookie(&sessionCookie)
			reqAdmin.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36")
			reqAdmin.Header.Set("Host", "localhost")
			reqAdmin.Host = "localhost"
			router.ServeHTTP(w, reqAdmin)
			if w.Code != 200 {
				b.Log(w.Body)
				b.Fatal("HTTP Error!")
			}

			{
				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", path, bytes.NewReader(nil))
				req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36")
				req.Header.Set("Host", "localhost")
				req.Host = "localhost"
				router.ServeHTTP(w, req)
				if w.Code != 200 {
					b.Log(w.Body)
					b.Fatal("HTTP Error!")
				}
			}
		}
	})

	cfg.Restore()
}

func BenchmarkTopicGuestAdminRouteParallelWithRouterPre2(b *testing.B) {
	runtime.GC()
}

func BenchmarkTopicGuestAdminRouteParallelWithRouterGC(b *testing.B) {
	binit(b)
	router, e := NewGenRouter(rcfg())
	if e != nil {
		b.Fatal(e)
	}
	cfg := NewStashConfig()
	c.Dev.DebugMode = false
	c.Dev.SuperDebug = false

	admin, e := c.Users.Get(1)
	if e != nil {
		b.Fatal(e)
	}
	if !admin.IsAdmin {
		b.Fatal("UID1 is not an admin")
	}
	uidCookie := http.Cookie{Name: "uid", Value: "1", Path: "/", MaxAge: c.Year}
	sessionCookie := http.Cookie{Name: "session", Value: admin.Session, Path: "/", MaxAge: c.Year}
	path := "/topic/hm." + benchTid
	//runtime.GC()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			reqAdmin := httptest.NewRequest("get", path, bytes.NewReader(nil))
			reqAdmin.AddCookie(&uidCookie)
			reqAdmin.AddCookie(&sessionCookie)
			reqAdmin.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36")
			reqAdmin.Header.Set("Host", "localhost")
			reqAdmin.Host = "localhost"
			router.ServeHTTP(w, reqAdmin)
			if w.Code != 200 {
				b.Log(w.Body)
				b.Fatal("HTTP Error!")
			}

			{
				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", path, bytes.NewReader(nil))
				req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36")
				req.Header.Set("Host", "localhost")
				req.Host = "localhost"
				router.ServeHTTP(w, req)
				if w.Code != 200 {
					b.Log(w.Body)
					b.Fatal("HTTP Error!")
				}
			}
		}

		runtime.GC()
	})

	cfg.Restore()
}

func BenchmarkTopicGuestRouteParallel(b *testing.B) {
	binit(b)
	cfg := NewStashConfig()
	c.Dev.DebugMode = false
	c.Dev.SuperDebug = false

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("get", "/topic/hm."+benchTid, bytes.NewReader(nil))
			user := c.GuestUser

			head, e := c.UserCheck(w, req, &user)
			if e != nil {
				b.Fatal(e)
			}
			//w.Body.Reset()
			routes.ViewTopic(w, req, &user, head, "1")
			if w.Code != 200 {
				b.Log(w.Body)
				b.Fatal("HTTP Error!")
			}
		}
	})
	cfg.Restore()
}

//before

func BenchmarkForumsRouteAdminParallelWithRouterGC2Pre(b *testing.B) {
	runtime.GC()
}

func BenchmarkForumsRouteAdminParallelWithRouterGC2(b *testing.B) {
	binit(b)
	router, e := NewGenRouter(rcfg())
	if e != nil {
		b.Fatal(e)
	}
	cfg := NewStashConfig()
	c.Dev.DebugMode = false
	c.Dev.SuperDebug = false

	admin, e := c.Users.Get(1)
	if e != nil {
		b.Fatal(e)
	}
	if !admin.IsAdmin {
		b.Fatal("UID1 is not an admin")
	}
	uidCookie := http.Cookie{Name: "uid", Value: "1", Path: "/", MaxAge: c.Year}
	sessionCookie := http.Cookie{Name: "session", Value: admin.Session, Path: "/", MaxAge: c.Year}
	path := "/forums/"
	//runtime.GC()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			reqAdmin := httptest.NewRequest("get", path, bytes.NewReader(nil))
			reqAdmin.AddCookie(&uidCookie)
			reqAdmin.AddCookie(&sessionCookie)
			reqAdmin.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36")
			reqAdmin.Header.Set("Host", "localhost")
			reqAdmin.Host = "localhost"
			router.ServeHTTP(w, reqAdmin)
			if w.Code != 200 {
				b.Log(w.Body)
				b.Fatal("HTTP Error!")
			}
		}

		runtime.GC()
	})

	cfg.Restore()
}

func BenchmarkForumsRouteAdminParallelWithRouterGCBrotliPre(b *testing.B) {
	runtime.GC()
}

func BenchmarkForumsRouteAdminParallelWithRouterGCBrotli(b *testing.B) {
	binit(b)
	router, e := NewGenRouter(rcfg())
	if e != nil {
		b.Fatal(e)
	}
	cfg := NewStashConfig()
	c.Dev.DebugMode = false
	c.Dev.SuperDebug = false

	admin, e := c.Users.Get(1)
	if e != nil {
		b.Fatal(e)
	}
	if !admin.IsAdmin {
		b.Fatal("UID1 is not an admin")
	}
	uidCookie := http.Cookie{Name: "uid", Value: "1", Path: "/", MaxAge: c.Year}
	sessionCookie := http.Cookie{Name: "session", Value: admin.Session, Path: "/", MaxAge: c.Year}
	path := "/forums/"
	//runtime.GC()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			reqAdmin := httptest.NewRequest("get", path, bytes.NewReader(nil))
			reqAdmin.AddCookie(&uidCookie)
			reqAdmin.AddCookie(&sessionCookie)
			reqAdmin.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36")
			reqAdmin.Header.Set("Accept-Encoding", "br")
			reqAdmin.Header.Set("Host", "localhost")
			reqAdmin.Host = "localhost"
			router.ServeHTTP(w, reqAdmin)
			if w.Code != 200 {
				b.Log(w.Body)
				b.Fatal("HTTP Error!")
			}
		}

		runtime.GC()
	})

	cfg.Restore()
}

//end
//before

func BenchmarkTopicRouteAdminParallelWithRouterGC2Pre(b *testing.B) {
	runtime.GC()
}

func BenchmarkTopicRouteAdminParallelWithRouterGC2(b *testing.B) {
	binit(b)
	router, e := NewGenRouter(rcfg())
	if e != nil {
		b.Fatal(e)
	}
	cfg := NewStashConfig()
	c.Dev.DebugMode = false
	c.Dev.SuperDebug = false

	admin, e := c.Users.Get(1)
	if e != nil {
		b.Fatal(e)
	}
	if !admin.IsAdmin {
		b.Fatal("UID1 is not an admin")
	}
	uidCookie := http.Cookie{Name: "uid", Value: "1", Path: "/", MaxAge: c.Year}
	sessionCookie := http.Cookie{Name: "session", Value: admin.Session, Path: "/", MaxAge: c.Year}
	path := "/topic/1"
	//runtime.GC()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			reqAdmin := httptest.NewRequest("get", path, bytes.NewReader(nil))
			reqAdmin.AddCookie(&uidCookie)
			reqAdmin.AddCookie(&sessionCookie)
			reqAdmin.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36")
			reqAdmin.Header.Set("Host", "localhost")
			reqAdmin.Host = "localhost"
			router.ServeHTTP(w, reqAdmin)
			if w.Code != 200 {
				b.Log(w.Body)
				b.Fatal("HTTP Error!")
			}
		}

		runtime.GC()
	})

	cfg.Restore()
}

func BenchmarkTopicRouteAdminParallelWithRouterGCBrotliPre(b *testing.B) {
	runtime.GC()
}

func BenchmarkTopicRouteAdminParallelWithRouterGCBrotli(b *testing.B) {
	binit(b)
	router, e := NewGenRouter(rcfg())
	if e != nil {
		b.Fatal(e)
	}
	cfg := NewStashConfig()
	c.Dev.DebugMode = false
	c.Dev.SuperDebug = false

	admin, e := c.Users.Get(1)
	if e != nil {
		b.Fatal(e)
	}
	if !admin.IsAdmin {
		b.Fatal("UID1 is not an admin")
	}
	uidCookie := http.Cookie{Name: "uid", Value: "1", Path: "/", MaxAge: c.Year}
	sessionCookie := http.Cookie{Name: "session", Value: admin.Session, Path: "/", MaxAge: c.Year}
	path := "/topic/1"
	//runtime.GC()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			reqAdmin := httptest.NewRequest("get", path, bytes.NewReader(nil))
			reqAdmin.AddCookie(&uidCookie)
			reqAdmin.AddCookie(&sessionCookie)
			reqAdmin.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36")
			reqAdmin.Header.Set("Accept-Encoding", "br")
			reqAdmin.Header.Set("Host", "localhost")
			reqAdmin.Host = "localhost"
			router.ServeHTTP(w, reqAdmin)
			if w.Code != 200 {
				b.Log(w.Body)
				b.Fatal("HTTP Error!")
			}
		}

		runtime.GC()
	})

	cfg.Restore()
}

//end

func BenchmarkTopicsRouteAdminParallelWithRouterGC2Pre(b *testing.B) {
	runtime.GC()
}

func BenchmarkTopicsRouteAdminParallelWithRouterGC2(b *testing.B) {
	binit(b)
	router, e := NewGenRouter(rcfg())
	if e != nil {
		b.Fatal(e)
	}
	cfg := NewStashConfig()
	c.Dev.DebugMode = false
	c.Dev.SuperDebug = false

	admin, e := c.Users.Get(1)
	if e != nil {
		b.Fatal(e)
	}
	if !admin.IsAdmin {
		b.Fatal("UID1 is not an admin")
	}
	uidCookie := http.Cookie{Name: "uid", Value: "1", Path: "/", MaxAge: c.Year}
	sessionCookie := http.Cookie{Name: "session", Value: admin.Session, Path: "/", MaxAge: c.Year}
	path := "/topics/"
	//runtime.GC()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			reqAdmin := httptest.NewRequest("get", path, bytes.NewReader(nil))
			reqAdmin.AddCookie(&uidCookie)
			reqAdmin.AddCookie(&sessionCookie)
			reqAdmin.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36")
			reqAdmin.Header.Set("Host", "localhost")
			reqAdmin.Host = "localhost"
			router.ServeHTTP(w, reqAdmin)
			if w.Code != 200 {
				b.Log(w.Body)
				b.Fatal("HTTP Error!")
			}
		}

		runtime.GC()
	})

	cfg.Restore()
}

func BenchmarkTopicsRouteAdminParallelWithRouterGCBrotliPre(b *testing.B) {
	runtime.GC()
}

func BenchmarkTopicsRouteAdminParallelWithRouterGCBrotli(b *testing.B) {
	binit(b)
	router, e := NewGenRouter(rcfg())
	if e != nil {
		b.Fatal(e)
	}
	cfg := NewStashConfig()
	c.Dev.DebugMode = false
	c.Dev.SuperDebug = false

	admin, e := c.Users.Get(1)
	if e != nil {
		b.Fatal(e)
	}
	if !admin.IsAdmin {
		b.Fatal("UID1 is not an admin")
	}
	uidCookie := http.Cookie{Name: "uid", Value: "1", Path: "/", MaxAge: c.Year}
	sessionCookie := http.Cookie{Name: "session", Value: admin.Session, Path: "/", MaxAge: c.Year}
	path := "/topics/"
	//runtime.GC()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			reqAdmin := httptest.NewRequest("get", path, bytes.NewReader(nil))
			reqAdmin.AddCookie(&uidCookie)
			reqAdmin.AddCookie(&sessionCookie)
			reqAdmin.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36")
			reqAdmin.Header.Set("Accept-Encoding", "br")
			reqAdmin.Header.Set("Host", "localhost")
			reqAdmin.Host = "localhost"
			router.ServeHTTP(w, reqAdmin)
			if w.Code != 200 {
				b.Log(w.Body)
				b.Fatal("HTTP Error!")
			}
		}

		runtime.GC()
	})

	cfg.Restore()
}

func BenchmarkTopicsRouteAdminParallelWithRouterGCGzipPre(b *testing.B) {
	runtime.GC()
}

func BenchmarkTopicsRouteAdminParallelWithRouterGCGzip(b *testing.B) {
	binit(b)
	router, e := NewGenRouter(rcfg())
	if e != nil {
		b.Fatal(e)
	}
	cfg := NewStashConfig()
	c.Dev.DebugMode = false
	c.Dev.SuperDebug = false

	admin, e := c.Users.Get(1)
	if e != nil {
		b.Fatal(e)
	}
	if !admin.IsAdmin {
		b.Fatal("UID1 is not an admin")
	}
	uidCookie := http.Cookie{Name: "uid", Value: "1", Path: "/", MaxAge: c.Year}
	sessionCookie := http.Cookie{Name: "session", Value: admin.Session, Path: "/", MaxAge: c.Year}
	path := "/topics/"
	//runtime.GC()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			reqAdmin := httptest.NewRequest("get", path, bytes.NewReader(nil))
			reqAdmin.AddCookie(&uidCookie)
			reqAdmin.AddCookie(&sessionCookie)
			reqAdmin.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36")
			reqAdmin.Header.Set("Accept-Encoding", "gzip")
			reqAdmin.Header.Set("Host", "localhost")
			reqAdmin.Host = "localhost"
			router.ServeHTTP(w, reqAdmin)
			if w.Code != 200 {
				b.Log(w.Body)
				b.Fatal("HTTP Error!")
			}
		}

		runtime.GC()
	})

	cfg.Restore()
}

func BenchmarkTopicGuestRouteParallelDebugMode(b *testing.B) {
	binit(b)
	cfg := NewStashConfig()
	c.Dev.DebugMode = true
	c.Dev.SuperDebug = false

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("get", "/topic/hm."+benchTid, bytes.NewReader(nil))
			user := c.GuestUser

			head, err := c.UserCheck(w, req, &user)
			if err != nil {
				b.Fatal(err)
			}
			//w.Body.Reset()
			routes.ViewTopic(w, req, &user, head, "1")
			if w.Code != 200 {
				b.Log(w.Body)
				b.Fatal("HTTP Error!")
			}
		}
	})
	cfg.Restore()
}

func BenchmarkAlertsRouteAdminParallelWithRouterGCPre(b *testing.B) {
	runtime.GC()
}

func BenchmarkAlertsRouteAdminParallelWithRouterGC(b *testing.B) {
	binit(b)
	router, e := NewGenRouter(rcfg())
	if e != nil {
		b.Fatal(e)
	}
	cfg := NewStashConfig()
	c.Dev.DebugMode = false
	c.Dev.SuperDebug = false

	admin, e := c.Users.Get(1)
	if e != nil {
		b.Fatal(e)
	}
	if !admin.IsAdmin {
		b.Fatal("UID1 is not an admin")
	}
	uidCookie := http.Cookie{Name: "uid", Value: "1", Path: "/", MaxAge: c.Year}
	sessionCookie := http.Cookie{Name: "session", Value: admin.Session, Path: "/", MaxAge: c.Year}
	path := "/api/?m=alerts"
	//runtime.GC()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			reqAdmin := httptest.NewRequest("get", path, bytes.NewReader(nil))
			reqAdmin.AddCookie(&uidCookie)
			reqAdmin.AddCookie(&sessionCookie)
			reqAdmin.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36")
			reqAdmin.Header.Set("Host", "localhost")
			reqAdmin.Host = "localhost"
			router.ServeHTTP(w, reqAdmin)
			if w.Code != 200 {
				b.Log(w.Body)
				b.Fatal("HTTP Error!")
			}
		}

		runtime.GC()
	})

	cfg.Restore()
}

func obRoute(b *testing.B, path string) {
	binit(b)
	cfg := NewStashConfig()
	c.Dev.DebugMode = false
	c.Dev.SuperDebug = false
	b.RunParallel(benchRoute(b, path))
	cfg.Restore()
}

func obRouteNoError(b *testing.B, path string) {
	binit(b)
	cfg := NewStashConfig()
	c.Dev.DebugMode = false
	c.Dev.SuperDebug = false
	b.RunParallel(benchRouteNoError(b, path))
	cfg.Restore()
}

func BenchmarkTopicsGuestRouteParallelWithRouter(b *testing.B) {
	obRoute(b, "/topics/")
}

func BenchmarkTopicsGuestJSRouteParallelWithRouter(b *testing.B) {
	obRoute(b, "/topics/?js=1")
}

func BenchmarkForumsGuestRouteParallelWithRouter(b *testing.B) {
	obRoute(b, "/forums/")
}

func BenchmarkForumGuestRouteParallelWithRouter(b *testing.B) {
	obRoute(b, "/forum/general.2")
}

func BenchmarkTopicGuestRouteParallelWithRouter(b *testing.B) {
	obRoute(b, "/topic/hm."+benchTid)
}

func BenchmarkTopicGuestRouteParallelWithRouterAlt(b *testing.B) {
	obRoute(b, "/topic/hm."+benchTid)
}

func BenchmarkBadRouteGuestRouteParallelWithRouter(b *testing.B) {
	obRouteNoError(b, "/garble/haa")
}

func BenchmarkAlertsRouteGuestParallelWithRouter(b *testing.B) {
	obRoute(b, "/api/?m=alerts")
}

// TODO: Alternate between member and guest to bust some CPU caches?

func binit(b *testing.B) {
	b.ReportAllocs()
	if err := gloinit(); err != nil {
		b.Fatal(err)
	}
}

type StashConfig struct {
	prev  bool
	prev2 bool
}

func NewStashConfig() *StashConfig {
	prev := c.Dev.DebugMode
	prev2 := c.Dev.SuperDebug
	return &StashConfig{prev, prev2}
}

func (cfg *StashConfig) Restore() {
	c.Dev.DebugMode = cfg.prev
	c.Dev.SuperDebug = cfg.prev2
}

func benchRoute(b *testing.B, path string) func(*testing.PB) {
	router, e := NewGenRouter(rcfg())
	if e != nil {
		b.Fatal(e)
	}
	return func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", path, bytes.NewReader(nil))
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36")
			req.Header.Set("Host", "localhost")
			req.Host = "localhost"
			router.ServeHTTP(w, req)
			if w.Code != 200 {
				b.Log(w.Body)
				b.Fatal("HTTP Error!")
			}
		}
	}
}

func benchRouteNoError(b *testing.B, path string) func(*testing.PB) {
	router, e := NewGenRouter(rcfg())
	if e != nil {
		b.Fatal(e)
	}
	return func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", path, bytes.NewReader(nil))
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36")
			req.Header.Set("Host", "localhost")
			req.Host = "localhost"
			router.ServeHTTP(w, req)
		}
	}
}

func BenchmarkProfileGuestRouteParallelWithRouter(b *testing.B) {
	obRoute(b, "/profile/admin.1")
}

func BenchmarkPopulateTopicWithRouter(b *testing.B) {
	b.ReportAllocs()
	topic, err := c.Topics.Get(benchTidI)
	if err != nil {
		debug.PrintStack()
		b.Fatal(err)
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < 25; i++ {
				_, err := c.Rstore.Create(topic, "hiii", "", 1)
				if err != nil {
					debug.PrintStack()
					b.Fatal(err)
				}
			}
		}
	})
}

//var fullPage = false

func BenchmarkTopicAdminFullPageRouteParallelWithRouter(b *testing.B) {
	/*if !fullPage {
		topic, err := c.Topics.Get(benchTidI)
		panicIfErr(err)
		for i := 0; i < 25; i++ {
			_, err = c.Rstore.Create(topic, "hiii", "::1", 1)
			panicIfErr(err)
		}
		fullPage = true
	}*/
	BenchmarkTopicAdminRouteParallel(b)
}

func BenchmarkTopicGuestFullPageRouteParallelWithRouter(b *testing.B) {
	/*if !fullPage {
		topic, err := c.Topics.Get(benchTidI)
		panicIfErr(err)
		for i := 0; i < 25; i++ {
			_, err = c.Rstore.Create(topic, "hiii", "::1", 1)
			panicIfErr(err)
		}
		fullPage = true
	}*/
	obRoute(b, "/topic/hm."+benchTid)
}

var benchTidI2 int
var benchTid2 string

func BenchmarkPopulateTopicMentionWithRouter(b *testing.B) {
	b.ReportAllocs()
	tid, err := c.Topics.Create(2, "test topic", "@1", 1, "")
	if err != nil {
		debug.PrintStack()
		b.Fatal(err)
	}
	benchTidI2 = tid
	benchTid2 = strconv.Itoa(tid)
	topic, err := c.Topics.Get(tid)
	if err != nil {
		debug.PrintStack()
		b.Fatal(err)
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < 25; i++ {
				_, err := c.Rstore.Create(topic, "@1", "", 1)
				if err != nil {
					debug.PrintStack()
					b.Fatal(err)
				}
			}
		}
	})
}

func BenchmarkTopicMentionAdminFullPageRouteParallelWithRouter(b *testing.B) {
	tI := benchTidI
	t := benchTid
	benchTidI = benchTidI2
	benchTid = benchTid2
	BenchmarkTopicAdminRouteParallel(b)
	benchTidI = tI
	benchTid = t
}

func BenchmarkTopicMentionGuestFullPageRouteParallelWithRouter(b *testing.B) {
	obRoute(b, "/topic/hm."+benchTid2)
}

var benchTidI3 int
var benchTid3 string

func BenchmarkPopulateTopic10MentionWithRouter(b *testing.B) {
	b.ReportAllocs()
	tid, err := c.Topics.Create(2, "test topic", "@1 @1 @1 @1 @1 @1 @1 @1 @1 @1", 1, "")
	if err != nil {
		debug.PrintStack()
		b.Fatal(err)
	}
	benchTidI3 = tid
	benchTid3 = strconv.Itoa(tid)
	topic, err := c.Topics.Get(tid)
	if err != nil {
		debug.PrintStack()
		b.Fatal(err)
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < 25; i++ {
				_, err := c.Rstore.Create(topic, "@1 @1 @1 @1 @1 @1 @1 @1 @1 @1", "", 1)
				if err != nil {
					debug.PrintStack()
					b.Fatal(err)
				}
			}
		}
	})
}

func BenchmarkTopic10MentionAdminFullPageRouteParallelWithRouter(b *testing.B) {
	tI := benchTidI
	t := benchTid
	benchTidI = benchTidI3
	benchTid = benchTid3
	BenchmarkTopicAdminRouteParallel(b)
	benchTidI = tI
	benchTid = t
}

func BenchmarkTopic10MentionGuestFullPageRouteParallelWithRouter(b *testing.B) {
	obRoute(b, "/topic/hm."+benchTid3)
}

var benchTidI4 int
var benchTid4 string

func BenchmarkPopulateTopic1ReplyWithRouter(b *testing.B) {
	b.ReportAllocs()
	tid, err := c.Topics.Create(2, "test topic", "hiii", 1, "")
	if err != nil {
		debug.PrintStack()
		b.Fatal(err)
	}
	benchTidI4 = tid
	benchTid4 = strconv.Itoa(tid)
	topic, err := c.Topics.Get(tid)
	if err != nil {
		debug.PrintStack()
		b.Fatal(err)
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := c.Rstore.Create(topic, "hiii", "", 1)
			if err != nil {
				debug.PrintStack()
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkTopic1ReplyAdminRouteParallelWithRouter(b *testing.B) {
	tI := benchTidI
	t := benchTid
	benchTidI = benchTidI4
	benchTid = benchTid4
	BenchmarkTopicAdminRouteParallel(b)
	benchTidI = tI
	benchTid = t
}

func BenchmarkTopic1ReplyGuestRouteParallelWithRouter(b *testing.B) {
	obRoute(b, "/topic/hm."+benchTid4)
}

var benchTidI5 int
var benchTid5 string
var benchUidI int

func BenchmarkPopulateTopic2UserWithRouter(b *testing.B) {
	b.ReportAllocs()
	nUid, err := c.Users.Create("testing", "testpass", "", 2, true)
	if err != nil {
		debug.PrintStack()
		b.Fatal(err)
	}
	benchUidI = nUid
	tid, err := c.Topics.Create(2, "test topic", "hiii", 1, "")
	if err != nil {
		debug.PrintStack()
		b.Fatal(err)
	}
	benchTidI5 = tid
	benchTid5 = strconv.Itoa(tid)
	topic, err := c.Topics.Get(tid)
	if err != nil {
		debug.PrintStack()
		b.Fatal(err)
	}
	b.RunParallel(func(pb *testing.PB) {
		var uid int
		for pb.Next() {
			for i := 0; i < 25; i++ {
				if i%2 == 0 {
					uid = nUid
				} else {
					uid = 1
				}
				_, err := c.Rstore.Create(topic, "hiii", "", uid)
				if err != nil {
					debug.PrintStack()
					b.Fatal(err)
				}
			}
		}
	})
}

func BenchmarkTopic2UserAdminFullPageRouteParallelWithRouter(b *testing.B) {
	tI := benchTidI
	t := benchTid
	benchTidI = benchTidI5
	benchTid = benchTid5
	BenchmarkTopicAdminRouteParallel(b)
	benchTidI = tI
	benchTid = t
}

func BenchmarkTopic2UserGuestFullPageRouteParallelWithRouter(b *testing.B) {
	obRoute(b, "/topic/hm."+benchTid5)
}

var benchTidI6 int
var benchTid6 string

func BenchmarkPopulateTopic3UserWithRouter(b *testing.B) {
	b.ReportAllocs()
	nUid, err := c.Users.Create("testing2", "testpass", "", 2, true)
	if err != nil {
		debug.PrintStack()
		b.Fatal(err)
	}
	tid, err := c.Topics.Create(2, "test topic", "hiii", 1, "")
	if err != nil {
		debug.PrintStack()
		b.Fatal(err)
	}
	benchTidI6 = tid
	benchTid6 = strconv.Itoa(tid)
	t, err := c.Topics.Get(tid)
	if err != nil {
		debug.PrintStack()
		b.Fatal(err)
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < 5; i++ {
				_, err := c.Rstore.Create(t, "hiii", "", 1)
				if err != nil {
					debug.PrintStack()
					b.Fatal(err)
				}
				_, err = c.Rstore.Create(t, "hiii", "", benchUidI)
				if err != nil {
					debug.PrintStack()
					b.Fatal(err)
				}
				_, err = c.Rstore.Create(t, "hiii", "", nUid)
				if err != nil {
					debug.PrintStack()
					b.Fatal(err)
				}
			}
		}
	})
}

func BenchmarkTopic3UserAdminFullPageRouteParallelWithRouter(b *testing.B) {
	tI := benchTidI
	t := benchTid
	benchTidI = benchTidI6
	benchTid = benchTid6
	BenchmarkTopicAdminRouteParallel(b)
	benchTidI = tI
	benchTid = t
}

func BenchmarkTopic3UserGuestFullPageRouteParallelWithRouter(b *testing.B) {
	obRoute(b, "/topic/hm."+benchTid6)
}

// TODO: Add topic poll page bench

// TODO: Make these routes compatible with the changes to the router
/*
func BenchmarkForumsAdminRouteParallel(b *testing.B) {
	b.ReportAllocs()
	gloinit()
	b.RunParallel(func(pb *testing.PB) {
		admin, err := users.Get(1)
		if err != nil {
			panic(err)
		}
		if !admin.Is_Admin {
			panic("UID1 is not an admin")
		}
		adminUidCookie := http.Cookie{Name:"uid",Value:"1",Path:"/",MaxAge: year}
		adminSessionCookie := http.Cookie{Name:"session",Value: admin.Session,Path:"/",MaxAge: year}

		forumsW := httptest.NewRecorder()
		forumsReq := httptest.NewRequest("get","/forums/",bytes.NewReader(nil))
		forumsReqAdmin := forums_req
		forumsReqAdmin.AddCookie(&adminUidCookie)
		forumsReqAdmin.AddCookie(&adminSessionCookie)
		forumsHandler := http.HandlerFunc(route_forums)

		for pb.Next() {
			forumsW.Body.Reset()
			forumsHandler.ServeHTTP(forumsW,forumsReqAdmin)
		}
	})
}

func BenchmarkForumsAdminRouteParallelProf(b *testing.B) {
	b.ReportAllocs()
	gloinit()

	b.RunParallel(func(pb *testing.PB) {
		admin, err := users.Get(1)
		if err != nil {
			panic(err)
		}
		if !admin.Is_Admin {
			panic("UID1 is not an admin")
		}
		adminUidCookie := http.Cookie{Name:"uid",Value:"1",Path:"/",MaxAge: year}
		adminSessionCookie := http.Cookie{Name:"session",Value: admin.Session,Path: "/",MaxAge: year}

		forumsW := httptest.NewRecorder()
		forumsReq := httptest.NewRequest("get","/forums/",bytes.NewReader(nil))
		forumsReqAdmin := forumsReq
		forumsReqAdmin.AddCookie(&admin_uid_cookie)
		forumsReqAdmin.AddCookie(&admin_session_cookie)
		forumsHandler := http.HandlerFunc(route_forums)
		f, err := os.Create("cpu_forums_admin_parallel.prof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		for pb.Next() {
			forumsW.Body.Reset()
			forumsHandler.ServeHTTP(forumsW,forumsReqAdmin)
		}
		pprof.StopCPUProfile()
	})
}

func BenchmarkRoutesSerial(b *testing.B) {
	b.ReportAllocs()
	admin, err := users.Get(1)
	if err != nil {
		panic(err)
	}
	if !admin.Is_Admin {
		panic("UID1 is not an admin")
	}

	admin_uid_cookie := http.Cookie{Name:"uid",Value:"1",Path:"/",MaxAge: year}
	admin_session_cookie := http.Cookie{Name:"session",Value: admin.Session,Path: "/",MaxAge: year}

	if plugins_inited {
		b.Log("Plugins have already been initialised, they can't be deinitialised so these tests will run with plugins on")
	}
	static_w := httptest.NewRecorder()
	static_req := httptest.NewRequest("get","/s/global.js",bytes.NewReader(nil))
	static_handler := http.HandlerFunc(route_static)

	topic_w := httptest.NewRecorder()
	topic_req := httptest.NewRequest("get","/topic/1",bytes.NewReader(nil))
	topic_req_admin := topic_req
	topic_req_admin.AddCookie(&admin_uid_cookie)
	topic_req_admin.AddCookie(&admin_session_cookie)
	topic_handler := http.HandlerFunc(route_topic_id)

	topics_w := httptest.NewRecorder()
	topics_req := httptest.NewRequest("get","/topics/",bytes.NewReader(nil))
	topics_req_admin := topics_req
	topics_req_admin.AddCookie(&admin_uid_cookie)
	topics_req_admin.AddCookie(&admin_session_cookie)
	topics_handler := http.HandlerFunc(route_topics)

	forum_w := httptest.NewRecorder()
	forum_req := httptest.NewRequest("get","/forum/1",bytes.NewReader(nil))
	forum_req_admin := forum_req
	forum_req_admin.AddCookie(&admin_uid_cookie)
	forum_req_admin.AddCookie(&admin_session_cookie)
	forum_handler := http.HandlerFunc(route_forum)

	forums_w := httptest.NewRecorder()
	forums_req := httptest.NewRequest("get","/forums/",bytes.NewReader(nil))
	forums_req_admin := forums_req
	forums_req_admin.AddCookie(&admin_uid_cookie)
	forums_req_admin.AddCookie(&admin_session_cookie)
	forums_handler := http.HandlerFunc(route_forums)

	gloinit()

	//f, err := os.Create("routes_bench_cpu.prof")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//pprof.StartCPUProfile(f)
	///defer pprof.StopCPUProfile()
	///pprof.StopCPUProfile()

	b.Run("static_recorder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//static_w.Code = 200
			static_w.Body.Reset()
			static_handler.ServeHTTP(static_w,static_req)
			//if static_w.Code != 200 {
			//	b.Print(static_w.Body)
			//	b.Fatal("HTTP Error!")
			//}
		}
	})

	b.Run("topic_admin_recorder", func(b *testing.B) {
		//f, err := os.Create("routes_bench_topic_cpu.prof")
		//if err != nil {
		//	b.Fatal(err)
		//}
		//pprof.StartCPUProfile(f)
		for i := 0; i < b.N; i++ {
			//topic_w.Code = 200
			topic_w.Body.Reset()
			topic_handler.ServeHTTP(topic_w,topic_req_admin)
			//if topic_w.Code != 200 {
			//	b.Print(topic_w.Body)
			//	b.Fatal("HTTP Error!")
			//}
		}
		//pprof.StopCPUProfile()
	})
	b.Run("topic_guest_recorder", func(b *testing.B) {
		f, err := os.Create("routes_bench_topic_cpu_2.prof")
		if err != nil {
			b.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		for i := 0; i < b.N; i++ {
			//topic_w.Code = 200
			topic_w.Body.Reset()
			topic_handler.ServeHTTP(topic_w,topic_req)
		}
		pprof.StopCPUProfile()
	})
	b.Run("topics_admin_recorder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//topics_w.Code = 200
			topics_w.Body.Reset()
			topics_handler.ServeHTTP(topics_w,topics_req_admin)
		}
	})
	b.Run("topics_guest_recorder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//topics_w.Code = 200
			topics_w.Body.Reset()
			topics_handler.ServeHTTP(topics_w,topics_req)
		}
	})
	b.Run("forum_admin_recorder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//forum_w.Code = 200
			forum_w.Body.Reset()
			forum_handler.ServeHTTP(forum_w,forum_req_admin)
		}
	})
	b.Run("forum_guest_recorder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//forum_w.Code = 200
			forum_w.Body.Reset()
			forum_handler.ServeHTTP(forum_w,forum_req)
		}
	})
	b.Run("forums_admin_recorder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//forums_w.Code = 200
			forums_w.Body.Reset()
			forums_handler.ServeHTTP(forums_w,forums_req_admin)
		}
	})
	b.Run("forums_guest_recorder", func(b *testing.B) {
		//f, err := os.Create("routes_bench_forums_cpu_2.prof")
		//if err != nil {
		//	b.Fatal(err)
		//}
		//pprof.StartCPUProfile(f)
		for i := 0; i < b.N; i++ {
			//forums_w.Code = 200
			forums_w.Body.Reset()
			forums_handler.ServeHTTP(forums_w,forums_req)
		}
		//pprof.StopCPUProfile()
	})

	if !plugins_inited {
		init_plugins()
	}

	b.Run("topic_admin_recorder_with_plugins", func(b *testing.B) {
		//f, err := os.Create("routes_bench_topic_cpu.prof")
		//if err != nil {
		//	b.Fatal(err)
		//}
		//pprof.StartCPUProfile(f)
		for i := 0; i < b.N; i++ {
			//topic_w.Code = 200
			topic_w.Body.Reset()
			topic_handler.ServeHTTP(topic_w,topic_req_admin)
			//if topic_w.Code != 200 {
			//	b.Print(topic_w.Body)
			//	b.Fatal("HTTP Error!")
			//}
		}
		//pprof.StopCPUProfile()
	})
	b.Run("topic_guest_recorder_with_plugins", func(b *testing.B) {
		//f, err := os.Create("routes_bench_topic_cpu_2.prof")
		//if err != nil {
		//	b.Fatal(err)
		//}
		//pprof.StartCPUProfile(f)
		for i := 0; i < b.N; i++ {
			//topic_w.Code = 200
			topic_w.Body.Reset()
			topic_handler.ServeHTTP(topic_w,topic_req)
		}
		//pprof.StopCPUProfile()
	})
	b.Run("topics_admin_recorder_with_plugins", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//topics_w.Code = 200
			topics_w.Body.Reset()
			topics_handler.ServeHTTP(topics_w,topics_req_admin)
		}
	})
	b.Run("topics_guest_recorder_with_plugins", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//topics_w.Code = 200
			topics_w.Body.Reset()
			topics_handler.ServeHTTP(topics_w,topics_req)
		}
	})
	b.Run("forum_admin_recorder_with_plugins", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//forum_w.Code = 200
			forum_w.Body.Reset()
			forum_handler.ServeHTTP(forum_w,forum_req_admin)
		}
	})
	b.Run("forum_guest_recorder_with_plugins", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//forum_w.Code = 200
			forum_w.Body.Reset()
			forum_handler.ServeHTTP(forum_w,forum_req)
		}
	})
	b.Run("forums_admin_recorder_with_plugins", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//forums_w.Code = 200
			forums_w.Body.Reset()
			forums_handler.ServeHTTP(forums_w,forums_req_admin)
		}
	})
	b.Run("forums_guest_recorder_with_plugins", func(b *testing.B) {
		//f, err := os.Create("routes_bench_forums_cpu_2.prof")
		//if err != nil {
		//	b.Fatal(err)
		//}
		//pprof.StartCPUProfile(f)
		for i := 0; i < b.N; i++ {
			//forums_w.Code = 200
			forums_w.Body.Reset()
			forums_handler.ServeHTTP(forums_w,forums_req)
		}
		//pprof.StopCPUProfile()
	})
}*/

func BenchmarkQueryTopicParallel(b *testing.B) {
	b.ReportAllocs()
	if err := gloinit(); err != nil {
		b.Fatal(err)
	}

	b.RunParallel(func(pb *testing.PB) {
		var tu c.TopicUser
		for pb.Next() {
			err := db.QueryRow("select topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, topics.ip, topics.views, topics.postCount, topics.likeCount, users.name, users.avatar, users.group, users.level from topics left join users ON topics.createdBy = users.uid where tid = ?", 1).Scan(&tu.Title, &tu.Content, &tu.CreatedBy, &tu.CreatedAt, &tu.IsClosed, &tu.Sticky, &tu.ParentID, &tu.IP, &tu.ViewCount, &tu.PostCount, &tu.LikeCount, &tu.CreatedByName, &tu.Avatar, &tu.Group, &tu.Level)
			if err == ErrNoRows {
				log.Fatal("No rows found!")
				return
			} else if err != nil {
				log.Fatal(err)
				return
			}
		}
	})
}

func BenchmarkQueryPreparedTopicParallel(b *testing.B) {
	b.ReportAllocs()
	if err := gloinit(); err != nil {
		b.Fatal(err)
	}

	b.RunParallel(func(pb *testing.PB) {
		var tu c.TopicUser

		getTopicUser, err := qgen.Builder.SimpleLeftJoin("topics", "users", "topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, topics.ip, topics.postCount, topics.likeCount, users.name, users.avatar, users.group, users.level", "topics.createdBy=users.uid", "tid=?", "", "")
		if err != nil {
			b.Fatal(err)
		}
		defer getTopicUser.Close()

		for pb.Next() {
			err := getTopicUser.QueryRow(1).Scan(&tu.Title, &tu.Content, &tu.CreatedBy, &tu.CreatedAt, &tu.IsClosed, &tu.Sticky, &tu.ParentID, &tu.IP, &tu.PostCount, &tu.LikeCount, &tu.CreatedByName, &tu.Avatar, &tu.Group, &tu.Level)
			if err == ErrNoRows {
				b.Fatal("No rows found!")
				return
			} else if err != nil {
				b.Fatal(err)
				return
			}
		}
	})
}

func BenchmarkUserGet(b *testing.B) {
	b.ReportAllocs()
	if err := gloinit(); err != nil {
		b.Fatal(err)
	}

	b.RunParallel(func(pb *testing.PB) {
		var err error
		for pb.Next() {
			_, err = c.Users.Get(1)
			if err != nil {
				b.Fatal(err)
				return
			}
		}
	})
}

func BenchmarkUserBypassGet(b *testing.B) {
	b.ReportAllocs()
	if err := gloinit(); err != nil {
		b.Fatal(err)
	}

	// Bypass the cache and always hit the database
	b.RunParallel(func(pb *testing.PB) {
		var err error
		for pb.Next() {
			_, err = c.Users.BypassGet(1)
			if err != nil {
				b.Fatal(err)
				return
			}
		}
	})
}

func BenchmarkQueriesSerial(b *testing.B) {
	b.ReportAllocs()
	var tu c.TopicUser
	b.Run("topic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := db.QueryRow("select topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, topics.ip, topics.postCount, topics.likeCount, users.name, users.avatar, users.group, users.level from topics left join users ON topics.createdBy = users.uid where tid = ?", 1).Scan(&tu.Title, &tu.Content, &tu.CreatedBy, &tu.CreatedAt, &tu.IsClosed, &tu.Sticky, &tu.ParentID, &tu.IP, &tu.PostCount, &tu.LikeCount, &tu.CreatedByName, &tu.Avatar, &tu.Group, &tu.Level)
			if err == ErrNoRows {
				b.Fatal("No rows found!")
				return
			} else if err != nil {
				b.Fatal(err)
				return
			}
		}
	})
	b.Run("topic_replies", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			rows, err := db.Query("select replies.rid, replies.content, replies.createdBy, replies.createdAt, replies.lastEdit, replies.lastEditBy, users.avatar, users.name, users.is_super_admin, users.group, users.level, replies.ip from replies left join users ON replies.createdBy = users.uid where tid = ?", 1)
			if err != nil {
				b.Fatal(err)
				return
			}
			defer rows.Close()

			for rows.Next() {
			}
			err = rows.Err()
			if err != nil {
				b.Fatal(err)
				return
			}
		}
	})

	var r c.ReplyUser
	var isSuperAdmin bool
	var group int
	b.Run("topic_replies_scan", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			rows, err := db.Query("select replies.rid, replies.content, replies.createdBy, replies.createdAt, replies.lastEdit, replies.lastEditBy, users.avatar, users.name, users.is_super_admin, users.group, users.level, replies.ip from replies left join users ON replies.createdBy = users.uid where tid = ?", 1)
			if err != nil {
				b.Fatal(err)
				return
			}
			for rows.Next() {
				err := rows.Scan(&r.ID, &r.Content, &r.CreatedBy, &r.CreatedAt, &r.LastEdit, &r.LastEditBy, &r.Avatar, &r.CreatedByName, &isSuperAdmin, &group, &r.Level, &r.IP)
				if err != nil {
					b.Fatal(err)
					return
				}
			}
			defer rows.Close()
			if err = rows.Err(); err != nil {
				b.Fatal(err)
				return
			}
		}
	})
}

// TODO: Take the attachment system into account in these parser benches
func BenchmarkParserSerial(b *testing.B) {
	if !c.PluginsInited {
		c.InitPlugins()
	}
	b.ReportAllocs()
	f := func(name, msg string) func(b *testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = c.ParseMessage(msg, 0, "", nil, nil)
			}
		}
	}
	f("empty_post", "")
	f("short_post", "Hey everyone, how's it going?")
	f("one_smily", "Hey everyone, how's it going? :)")
	f("five_smilies", "Hey everyone, how's it going? :):):):):)")
	f("ten_smilies", "Hey everyone, how's it going? :):):):):):):):):):)")
	f("twenty_smilies", "Hey everyone, how's it going? :):):):):):):):):):):):):):):):):):):):)")
}

func BenchmarkBBCodePluginWithRegexpSerial(b *testing.B) {
	if !c.PluginsInited {
		c.InitPlugins()
	}
	b.ReportAllocs()
	f := func(name string, msg string) {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = e.BbcodeRegexParse(msg)
			}
		})
	}
	f("empty_post", "")
	f("short_post", "Hey everyone, how's it going?")
	f("one_smily", "Hey everyone, how's it going? :)")
	f("five_smilies", "Hey everyone, how's it going? :):):):):)")
	f("ten_smilies", "Hey everyone, how's it going? :):):):):):):):):):)")
	f("twenty_smilies", "Hey everyone, how's it going? :):):):):):):):):):):):):):):):):):):):)")
	f("one_bold", "[b]H[/b]ey everyone, how's it going?")
	f("five_bold", "[b]H[/b][b]e[/b][b]y[/b] [b]e[/b][b]v[/b]eryone, how's it going?")
	f("ten_bold", "[b]H[/b][b]e[/b][b]y[/b] [b]e[/b][b]v[/b][b]e[/b][b]r[/b][b]y[/b][b]o[/b][b]n[/b]e, how's it going?")
}

func BenchmarkBBCodePluginWithoutCodeTagSerial(b *testing.B) {
	b.ReportAllocs()
	f := func(name string, msg string) {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = e.BbcodeParseWithoutCode(msg)
			}
		})
	}
	f("empty_post", "")
	f("short_post", "Hey everyone, how's it going?")
	f("one_smily", "Hey everyone, how's it going? :)")
	f("five_smilies", "Hey everyone, how's it going? :):):):):)")
	f("ten_smilies", "Hey everyone, how's it going? :):):):):):):):):):)")
	f("twenty_smilies", "Hey everyone, how's it going? :):):):):):):):):):):):):):):):):):):):)")
	f("one_bold", "[b]H[/b]ey everyone, how's it going?")
	f("five_bold", "[b]H[/b][b]e[/b][b]y[/b] [b]e[/b][b]v[/b]eryone, how's it going?")
	f("ten_bold", "[b]H[/b][b]e[/b][b]y[/b] [b]e[/b][b]v[/b][b]e[/b][b]r[/b][b]y[/b][b]o[/b][b]n[/b]e, how's it going?")
}

func BenchmarkBBCodePluginWithFullParserSerial(b *testing.B) {
	b.ReportAllocs()
	f := func(name string, msg string) {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = e.BbcodeFullParse(msg)
			}
		})
	}
	f("empty_post", "")
	f("short_post", "Hey everyone, how's it going?")
	f("one_smily", "Hey everyone, how's it going? :)")
	f("five_smilies", "Hey everyone, how's it going? :):):):):)")
	f("ten_smilies", "Hey everyone, how's it going? :):):):):):):):):):)")
	f("twenty_smilies", "Hey everyone, how's it going? :):):):):):):):):):):):):):):):):):):):)")
	f("one_bold", "[b]H[/b]ey everyone, how's it going?")
	f("five_bold", "[b]H[/b][b]e[/b][b]y[/b] [b]e[/b][b]v[/b]eryone, how's it going?")
	f("ten_bold", "[b]H[/b][b]e[/b][b]y[/b] [b]e[/b][b]v[/b][b]e[/b][b]r[/b][b]y[/b][b]o[/b][b]n[/b]e, how's it going?")
}

func TestLevels(t *testing.T) {
	levels := c.GetLevels(40)
	for level, score := range levels {
		sscore := strconv.FormatFloat(score, 'f', -1, 64)
		t.Log("Level: " + strconv.Itoa(level) + " Score: " + sscore)
	}
}

// TODO: Make this compatible with the changes to the router
/*
func TestStaticRoute(t *testing.T) {
	gloinit()
	if !plugins_inited {
		init_plugins()
	}

	static_w := httptest.NewRecorder()
	static_req := httptest.NewRequest("get","/s/global.js",bytes.NewReader(nil))
	static_handler := http.HandlerFunc(route_static)

	static_handler.ServeHTTP(static_w,static_req)
	if static_w.Code != 200 {
		t.Fatal(static_w.Body)
	}
}
*/

/*func TestTopicAdminRoute(t *testing.T) {
	gloinit()
	if !plugins_inited {
		init_plugins()
	}

	admin, err := users.Get(1)
	if err != nil {
		panic(err)
	}
	if !admin.Is_Admin {
		panic("UID1 is not an admin")
	}

	admin_uid_cookie := http.Cookie{Name:"uid",Value:"1",Path:"/",MaxAge: year}
	admin_session_cookie := http.Cookie{Name:"session",Value: admin.Session,Path:"/",MaxAge: year}

	topic_w := httptest.NewRecorder()
	topic_req := httptest.NewRequest("get","/topic/1",bytes.NewReader(nil))
	topic_req_admin := topic_req
	topic_req_admin.AddCookie(&admin_uid_cookie)
	topic_req_admin.AddCookie(&admin_session_cookie)
	topic_handler := http.HandlerFunc(route_topic_id)

	topic_handler.ServeHTTP(topic_w,topic_req_admin)
	if topic_w.Code != 200 {
		t.Print(topic_w.Body)
		t.Fatal("HTTP Error!")
	}
	t.Print("No problems found in the topic-admin route!")
}*/

/*func TestTopicGuestRoute(t *testing.T) {
	gloinit()
	if !plugins_inited {
		init_plugins()
	}

	topic_w := httptest.NewRecorder()
	topic_req := httptest.NewRequest("get","/topic/1",bytes.NewReader(nil))
	topic_handler := http.HandlerFunc(route_topic_id)

	topic_handler.ServeHTTP(topic_w,topic_req)
	if topic_w.Code != 200 {
		t.Print(topic_w.Body)
		t.Fatal("HTTP Error!")
	}
	t.Print("No problems found in the topic-guest route!")
}*/

// TODO: Make these routes compatible with the changes to the router
/*
func TestForumsAdminRoute(t *testing.T) {
	gloinit()
	if !plugins_inited {
		init_plugins()
	}

	admin, err := users.Get(1)
	if err != nil {
		t.Fatal(err)
	}
	if !admin.Is_Admin {
		t.Fatal("UID1 is not an admin")
	}
	adminUidCookie := http.Cookie{Name:"uid",Value:"1",Path:"/",MaxAge: year}
	adminSessionCookie := http.Cookie{Name:"session",Value: admin.Session,Path:"/",MaxAge: year}

	forumsW := httptest.NewRecorder()
	forumsReq := httptest.NewRequest("get","/forums/",bytes.NewReader(nil))
	forumsReqAdmin := forums_req
	forumsReqAdmin.AddCookie(&adminUidCookie)
	forumsReqAdmin.AddCookie(&adminSessionCookie)
	forumsHandler := http.HandlerFunc(route_forums)

	forumsHandler.ServeHTTP(forumsW,forumsReqAdmin)
	if forumsW.Code != 200 {
		t.Fatal(forumsW.Body)
	}
}

func TestForumsGuestRoute(t *testing.T) {
	gloinit()
	if !plugins_inited {
		init_plugins()
	}

	forums_w := httptest.NewRecorder()
	forums_req := httptest.NewRequest("get","/forums/",bytes.NewReader(nil))
	forums_handler := http.HandlerFunc(route_forums)

	forums_handler.ServeHTTP(forums_w,forums_req)
	if forums_w.Code != 200 {
		t.Fatal(forums_w.Body)
	}
}
*/

/*func TestForumAdminRoute(t *testing.T) {
	gloinit()
	if !plugins_inited {
		init_plugins()
	}

	admin, err := users.Get(1)
	if err != nil {
		panic(err)
	}
	if !admin.Is_Admin {
		panic("UID1 is not an admin")
	}
	admin_uid_cookie := http.Cookie{Name:"uid",Value:"1",Path:"/",MaxAge: year}
	admin_session_cookie := http.Cookie{Name:"session",Value: admin.Session,Path:"/",MaxAge: year}

	forum_w := httptest.NewRecorder()
	forum_req := httptest.NewRequest("get","/forum/1",bytes.NewReader(nil))
	forum_req_admin := forum_req
	forum_req_admin.AddCookie(&admin_uid_cookie)
	forum_req_admin.AddCookie(&admin_session_cookie)
	forum_handler := http.HandlerFunc(route_forum)

	forum_handler.ServeHTTP(forum_w,forum_req_admin)
	if forum_w.Code != 200 {
		t.Print(forum_w.Body)
		t.Fatal("HTTP Error!")
	}
}*/

/*func TestForumGuestRoute(t *testing.T) {
	gloinit()
	if !plugins_inited {
		init_plugins()
	}

	forum_w := httptest.NewRecorder()
	forum_req := httptest.NewRequest("get","/forum/2",bytes.NewReader(nil))
	forum_handler := http.HandlerFunc(route_forum)

	forum_handler.ServeHTTP(forum_w,forum_req)
	if forum_w.Code != 200 {
		t.Print(forum_w.Body)
		t.Fatal("HTTP Error!")
	}
}*/

/*func TestAlerts(t *testing.T) {
	gloinit()
	if !plugins_inited {
		init_plugins()
	}
	db = db_test
	alert_w := httptest.NewRecorder()
	alert_req := httptest.NewRequest("get","/api/?action=get&module=alerts&format=json",bytes.NewReader(nil))
	alert_handler := http.HandlerFunc(route_api)
	//testdb.StubQuery()
	testdb.SetQueryFunc(func(query string) (result sql.Rows, err error) {
		cols := []string{"asid","actor","targetUser","event","elementType","elementID"}
		rows := `1,1,0,like,post,5
		1,1,0,friend_invite,user,2`
		return testdb.RowsFromCSVString(cols,rows), nil
	})

	alert_handler.ServeHTTP(alert_w,alert_req)
	t.Print(alert_w.Body)
	if alert_w.Code != 200 {
		t.Fatal("HTTP Error!")
	}
	db = db_prod
}*/

func TestSplittyThing(t *testing.T) {
	var extraData string
	path := "/pages/hohoho"
	t.Log("Raw Path:", path)
	if path[len(path)-1] != '/' {
		extraData = path[strings.LastIndexByte(path, '/')+1:]
		path = path[:strings.LastIndexByte(path, '/')+1]
	}
	t.Log("Path:", path)
	t.Log("Extra Data:", extraData)
	t.Log("Path Bytes:", []byte(path))
	t.Log("Extra Data Bytes:", []byte(extraData))

	t.Log("Splitty thing test")
	path = "/topics/"
	extraData = ""
	t.Log("Raw Path:", path)
	if path[len(path)-1] != '/' {
		extraData = path[strings.LastIndexByte(path, '/')+1:]
		path = path[:strings.LastIndexByte(path, '/')+1]
	}
	t.Log("Path:", path)
	t.Log("Extra Data:", extraData)
	t.Log("Path Bytes:", []byte(path))
	t.Log("Extra Data Bytes:", []byte(extraData))
}
