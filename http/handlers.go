package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"stageai.tech/sunshine/sunshine"
	"stageai.tech/sunshine/sunshine/graphql"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/sentry"
	"stageai.tech/sunshine/sunshine/services"

	"github.com/getsentry/raven-go"
	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

const (
	uuidRe      = "{id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}"
	countHeader = "X-Documents-Count"
	filenameRe  = `{filename:.+}`
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

// New returns a new handler which a *http.ServeMux with applied
// middlewares.
func New(env *services.Env) http.Handler {
	var (
		auth   = NewAuth(env)
		authfp = newAuthfp(env)
		user   = newUser(env)
		org    = newOrg(env)
		asset  = newAsset(env)
		proj   = newProject(env)
		meet   = newMeeting(env)
		contr  = newContractHandler(env)
		stats  = newStats(env)
		wp     = newWP(env)
		mp     = newMP(env)
		gd     = newGDPR(env)
		gqlh   = graphql.Handler(env)
		fa     = newForfaitingApplication(env)
		mux    = mux.NewRouter().StrictSlash(true).UseEncodedPath()
	)

	mux.HandleFunc("/debug/ping", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "version=%s\ngo=%s\n", sunshine.Version(), runtime.Version())
	})
	mux.HandleFunc("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(basepath, "..", "openapi.json"))
	})
	mux.Handle("/", graphql.Playground("Sunshine API console", "/query"))
	mux.HandleFunc("/graphiql", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Del("Content-Type")
		graphql.Graphiql("/query")(w, r)
	})

	mux.Handle("/user", handlers.MethodHandler{
		"GET":  http.HandlerFunc(user.list),
		"POST": http.HandlerFunc(user.create),
	})
	mux.Handle("/user/"+uuidRe, handlers.MethodHandler{
		"GET": http.HandlerFunc(user.get),
		"PUT": http.HandlerFunc(user.update),
	})
	mux.Handle("/user/"+uuidRe+"/upload", handlers.MethodHandler{
		"POST": http.HandlerFunc(user.upload),
	})
	mux.Handle("/user/"+uuidRe+"/assets", handlers.MethodHandler{
		"GET": http.HandlerFunc(asset.list),
	})
	mux.Handle("/user/"+uuidRe+"/organizations", handlers.MethodHandler{
		"GET": http.HandlerFunc(org.list),
	})
	mux.Handle("/user/"+uuidRe+"/projects", handlers.MethodHandler{
		"GET": http.HandlerFunc(proj.list),
	})
	mux.Handle("/user/"+uuidRe+"/"+filenameRe, handlers.MethodHandler{
		"DELETE": http.HandlerFunc(user.delFile),
		"GET":    http.HandlerFunc(user.getFile),
		"HEAD":   http.HandlerFunc(user.getFile),
	})

	mux.Handle("/auth/login", handlers.MethodHandler{
		"POST": http.HandlerFunc(auth.login),
	})
	mux.Handle("/auth/change_password", handlers.MethodHandler{
		"POST": http.HandlerFunc(auth.changePassword),
	})
	mux.Handle("/auth/forgotten_password", handlers.MethodHandler{
		"POST": http.HandlerFunc(authfp.declare),
	})
	mux.Handle("/auth/forgotten_password/"+uuidRe, handlers.MethodHandler{
		"GET":  http.HandlerFunc(authfp.confirm),
		"POST": http.HandlerFunc(authfp.change),
	})
	mux.Handle("/auth/logout", handlers.MethodHandler{
		"GET": http.HandlerFunc(auth.Logout),
	})
	mux.Handle("/confirm_user/"+uuidRe, handlers.MethodHandler{
		"POST": http.HandlerFunc(auth.confirm),
	})

	mux.Handle("/organization", handlers.MethodHandler{
		"GET":  http.HandlerFunc(org.list),
		"POST": http.HandlerFunc(org.create),
	})
	mux.Handle("/organization/"+uuidRe, handlers.MethodHandler{
		"GET": http.HandlerFunc(org.get),
		"PUT": http.HandlerFunc(org.update),
	})
	mux.Handle("/organization/"+uuidRe+"/upload", handlers.MethodHandler{
		"POST": http.HandlerFunc(org.upload),
	})
	mux.Handle("/organization/"+uuidRe+"/roles", handlers.MethodHandler{
		"POST":   http.HandlerFunc(org.addRole),
		"DELETE": http.HandlerFunc(org.removeRole),
	})
	mux.Handle("/organization/"+uuidRe+"/meetings", handlers.MethodHandler{
		"GET": http.HandlerFunc(org.getMeetings),
	})
	mux.Handle("/organization/"+uuidRe+"/"+filenameRe, handlers.MethodHandler{
		"DELETE": http.HandlerFunc(org.delFile),
		"GET":    http.HandlerFunc(org.getFile),
		"HEAD":   http.HandlerFunc(org.getFile),
	})

	mux.Handle("/asset", handlers.MethodHandler{
		"GET":  http.HandlerFunc(asset.list),
		"POST": http.HandlerFunc(asset.create),
	})
	mux.Handle("/asset/"+uuidRe, handlers.MethodHandler{
		"GET": http.HandlerFunc(asset.get),
		"PUT": http.HandlerFunc(asset.update),
	})
	mux.Handle("/asset/"+uuidRe+"/upload", handlers.MethodHandler{
		"POST": http.HandlerFunc(asset.upload),
	})
	mux.Handle("/asset/"+uuidRe+"/"+filenameRe, handlers.MethodHandler{
		"DELETE": http.HandlerFunc(asset.delFile),
		"GET":    http.HandlerFunc(asset.getFile),
		"HEAD":   http.HandlerFunc(asset.getFile),
	})

	mux.Handle("/project", handlers.MethodHandler{
		"GET":  http.HandlerFunc(proj.list),
		"POST": http.HandlerFunc(proj.create),
	})
	mux.Handle("/project/reports", handlers.MethodHandler{
		"GET": http.HandlerFunc(proj.reports),
	})
	mux.Handle("/project/"+uuidRe, handlers.MethodHandler{
		"GET": http.HandlerFunc(proj.get),
		"PUT": http.HandlerFunc(proj.update),
	})
	mux.Handle("/project/"+uuidRe+"/upload", handlers.MethodHandler{
		"POST": http.HandlerFunc(proj.upload),
	})
	mux.Handle("/project/"+uuidRe+"/download/english",
		handlers.MethodHandler{
			"GET": http.HandlerFunc(contr.downloadEnglishPDF),
		},
	)
	mux.Handle("/project/"+uuidRe+"/download/native",
		handlers.MethodHandler{
			"GET": http.HandlerFunc(contr.downloadNativePDF),
		},
	)
	mux.Handle("/project/"+uuidRe+"/tex/english",
		handlers.MethodHandler{
			"GET": http.HandlerFunc(contr.downloadEnglishTeX),
		},
	)
	mux.Handle("/project/"+uuidRe+"/tex/native",
		handlers.MethodHandler{
			"GET": http.HandlerFunc(contr.downloadNativeTeX),
		},
	)
	mux.Handle("/project/"+uuidRe+"/agreement/download/native",
		handlers.MethodHandler{
			"GET": http.HandlerFunc(contr.downloadNativeAgreementPDF),
		},
	)
	mux.Handle("/project/"+uuidRe+"/agreement/download/english",
		handlers.MethodHandler{
			"GET": http.HandlerFunc(contr.downloadEnglishAgreementPDF),
		},
	)
	mux.Handle("/project/"+uuidRe+"/agreement/tex/native",
		handlers.MethodHandler{
			"GET": http.HandlerFunc(contr.downloadNativeAgreementTex),
		},
	)
	mux.Handle("/project/"+uuidRe+"/agreement/tex/english",
		handlers.MethodHandler{
			"GET": http.HandlerFunc(contr.downloadEnglishAgreementTex),
		},
	)
	mux.Handle("/project/"+uuidRe+"/agreement/fields",
		handlers.MethodHandler{
			"PUT": http.HandlerFunc(contr.updateAgreement),
			"GET": http.HandlerFunc(contr.getAgreement),
		},
	)
	mux.Handle("/project/"+uuidRe+"/fields",
		handlers.MethodHandler{
			"PUT": http.HandlerFunc(contr.updateFields),
			"GET": http.HandlerFunc(contr.getFields),
		},
	)
	mux.Handle("/project/"+uuidRe+"/maintenance/fields",
		handlers.MethodHandler{
			"PUT": http.HandlerFunc(contr.updateMaintenance),
			"GET": http.HandlerFunc(contr.getMaintenance),
		},
	)

	mux.Handle("/project/"+uuidRe+"/indoorclima", handlers.MethodHandler{
		"GET": http.HandlerFunc(contr.getIndoorClima),
		"PUT": http.HandlerFunc(contr.updateIndoorClima),
	})
	mux.Handle("/project/"+uuidRe+`/markdown`,
		handlers.MethodHandler{
			"GET": http.HandlerFunc(contr.getMarkdown),
			"PUT": http.HandlerFunc(contr.updateMarkdown),
		},
	)
	mux.Handle("/project/"+uuidRe+`/roles`,
		handlers.MethodHandler{
			"POST":   http.HandlerFunc(proj.addRole),
			"DELETE": http.HandlerFunc(proj.removeRole),
		},
	)
	mux.Handle("/project/"+uuidRe+"/meetings", handlers.MethodHandler{
		"GET": http.HandlerFunc(proj.getMeetings),
	})

	mux.Handle("/project/"+uuidRe+`/annex{n:[0-9]+}/{table:[0-9a-zA-Z_]+}`,
		handlers.MethodHandler{
			"GET": http.HandlerFunc(contr.getTable),
			"PUT": http.HandlerFunc(contr.updateTable),
		},
	)
	mux.Handle("/project/"+uuidRe+"/"+filenameRe, handlers.MethodHandler{
		"DELETE": http.HandlerFunc(proj.delFile),
		"GET":    http.HandlerFunc(proj.getFile),
		"HEAD":   http.HandlerFunc(proj.getFile),
	})

	mux.Handle("/gdpr/"+uuidRe+"/upload", handlers.MethodHandler{
		"POST": http.HandlerFunc(gd.upload),
	})
	mux.Handle("/gdpr/"+uuidRe+"/"+filenameRe, handlers.MethodHandler{
		"DELETE": http.HandlerFunc(gd.delFile),
		"GET":    http.HandlerFunc(gd.getFile),
		"HEAD":   http.HandlerFunc(gd.getFile),
	})

	mux.Handle("/country_stats", handlers.MethodHandler{
		"GET": http.HandlerFunc(stats.getCountryStats),
	})

	mux.Handle("/meeting/"+uuidRe+"/upload", handlers.MethodHandler{
		"POST": http.HandlerFunc(meet.upload),
	})
	mux.Handle("/meeting/"+uuidRe+"/"+filenameRe, handlers.MethodHandler{
		"DELETE": http.HandlerFunc(meet.delFile),
		"GET":    http.HandlerFunc(meet.getFile),
		"HEAD":   http.HandlerFunc(meet.getFile),
	})

	mux.Handle("/forfaitinga/"+uuidRe+"/upload", handlers.MethodHandler{
		"POST": http.HandlerFunc(fa.upload),
	})
	mux.Handle("/forfaitinga/"+uuidRe+"/"+filenameRe, handlers.MethodHandler{
		"DELETE": http.HandlerFunc(fa.delFile),
		"GET":    http.HandlerFunc(fa.getFile),
		"HEAD":   http.HandlerFunc(fa.getFile),
	})

	mux.Handle("/workphase/"+uuidRe, handlers.MethodHandler{
		"GET": http.HandlerFunc(wp.getWP),
	})
	mux.Handle("/workphase/"+uuidRe+"/upload", handlers.MethodHandler{
		"POST": http.HandlerFunc(wp.uploadWP),
	})
	mux.Handle("/workphase/"+uuidRe+"/"+filenameRe, handlers.MethodHandler{
		"DELETE": http.HandlerFunc(wp.delFileWP),
		"GET":    http.HandlerFunc(wp.getFileWP),
		"HEAD":   http.HandlerFunc(wp.getFileWP),
	})

	mux.Handle("/monitoringphase/"+uuidRe, handlers.MethodHandler{
		"GET": http.HandlerFunc(mp.getMP),
	})
	mux.Handle("/monitoringphase/"+uuidRe+"/upload", handlers.MethodHandler{
		"POST": http.HandlerFunc(mp.uploadMP),
	})
	mux.Handle("/monitoringphase/"+uuidRe+"/"+filenameRe, handlers.MethodHandler{
		"DELETE": http.HandlerFunc(mp.delFileMP),
		"GET":    http.HandlerFunc(mp.getFileMP),
		"HEAD":   http.HandlerFunc(mp.getFileMP),
	})

	mux.Handle("/query", gqlh)
	mux.Use(func(h http.Handler) http.Handler { return authMiddleware(h, env) })
	var h http.Handler = mux
	if len(env.General.AllowedOrigins) > 0 {
		h = handlers.CORS(
			handlers.AllowCredentials(),
			handlers.AllowedHeaders([]string{"Content-Type"}),
			handlers.AllowedOrigins(env.General.AllowedOrigins),
			handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"}),
			handlers.ExposedHeaders([]string{"X-Documents-Count"}),
		)(mux)
	}
	if !env.Debug {
		h = handlers.LoggingHandler(os.Stdout, h)
		h = raven.Recoverer(h)
	}
	return ContentTypeResponseHandler(h, "application/json")
}

func ContentTypeResponseHandler(h http.Handler, contentType string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		h.ServeHTTP(w, r)
	})
}

func authMiddleware(next http.Handler, env *services.Env) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := services.Session(env.SessionStore, r)
		ctx, ok := sessionContext(r.Context(), env, s)
		if !ok && len(s.Values) > 0 {
			// s.Values is not empty when session is invalid, so
			// let's empty it to avoid subsequent bad writes.
			sentry.Report(errors.New("bad session store state"),
				sentry.CaptureRequest(r),
				func(e sentry.Extra) { e["session"] = s.Values },
			)

			NewAuth(env).Logout(w, r)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func sessionContext(ctx context.Context, env *services.Env, s *sessions.Session) (context.Context, bool) {
	id, logged := s.Values["id"].(uuid.UUID)
	if !logged {
		return ctx, false
	}

	token, err := env.TokenStore.Get(ctx, models.SessionToken, id)
	if err != nil {
		// timed out or invalid token
		return ctx, false
	}

	return services.WithContext(ctx, token), true
}
