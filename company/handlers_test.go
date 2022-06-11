package company

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/julienschmidt/httprouter"
)

const defaultConnURL = "postgresql://xmusr:xmpass@localhost/xm_db"

func init() {
	if _, ok := os.LookupEnv("PGX_TEST_DATABASE"); !ok {
		os.Setenv("PGX_TEST_DATABASE", defaultConnURL)
	}
}

func setup(stmts []string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.Connect(context.Background(), os.Getenv("PGX_TEST_DATABASE"))
	if err != nil {
		return nil, err
	}

	tx, err := pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	for _, stmt := range stmts {
		if err == nil {
			_, err = tx.Exec(context.Background(), stmt)
		}
	}
	if err != nil {
		pool.Close()
		return nil, err
	}

	tx.Commit(context.Background())

	return pool, nil
}

func clean(pool *pgxpool.Pool, stmts []string) {
	defer pool.Close()

	tx, err := pool.Begin(context.Background())
	if err != nil {
		return
	}
	defer tx.Rollback(context.Background())

	for _, stmt := range stmts {
		if err == nil {
			_, err = tx.Exec(context.Background(), stmt)
		}
	}
	if err != nil {
		return
	}

	tx.Commit(context.Background())
}

func TestList(t *testing.T) {
	stmts := []string{
		"insert into company (name, code, country, website, phone) values ('Test0', 'TST0', 'TT', 'NA', '0')",
	}
	pool, err := setup(stmts)
	if err != nil {
		t.Fatal(err)
	}
	defer clean(pool, []string{
		"delete from company where name in ('Test0')",
	})

	model := NewModel(pool)

	tests := []struct {
		url  string
		want int
	}{
		{
			"/companies",
			http.StatusOK,
		},
		{
			"/companies?name=Test0",
			http.StatusOK,
		},
		{
			"/companies?name=Test0&code=TST0&country=TT&website=NA&phone=0",
			http.StatusOK,
		},
		{
			"/companies?name=Unknown",
			http.StatusBadRequest,
		},
		{
			"/companies?eman=nwonknU",
			http.StatusBadRequest,
		},
		{
			"/companies?name",
			http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		req, _ := http.NewRequest(http.MethodGet, test.url, nil)
		rr := httptest.NewRecorder()

		List(model).ServeHTTP(rr, req)

		if status := rr.Code; status != test.want {
			t.Errorf("%v %v status %v; want status %v", req.Method, test.url, status, test.want)
		}
	}
}

func TestFind(t *testing.T) {
	stmts := []string{
		"insert into company (id, name, code, country, website, phone) values (100, 'Test0', 'TST0', 'TT', 'NA', '0')",
	}
	pool, err := setup(stmts)
	if err != nil {
		t.Fatal(err)
	}
	defer clean(pool, []string{
		"delete from company where name in ('Test0')",
	})

	model := NewModel(pool)

	tests := []struct {
		url  string
		want int
	}{
		{
			"/companies/100",
			http.StatusOK,
		},
		{
			"/companies/200",
			http.StatusBadRequest,
		},
	}

	router := httprouter.New()
	router.GET("/companies/:id", Find(model))

	for _, test := range tests {
		req, _ := http.NewRequest(http.MethodGet, test.url, nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if status := rr.Code; status != test.want {
			t.Errorf("%v %v status %v; want status %v", req.Method, test.url, status, test.want)
		}
	}
}

func TestCreate(t *testing.T) {
	pool, err := setup(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer clean(pool, []string{
		"delete from company where name in ('Test0')",
		"delete from company where name is null",
	})

	model := NewModel(pool)

	tests := []struct {
		url  string
		body string
		want int
	}{
		{
			"/companies",
			`{
				"name": "Test0",
				"code": "TST0",
				"country": "TT",
				"website": "http",
				"phone": "0"
			}`,
			http.StatusOK,
		},
		{
			"/companies",
			`{
				"name": "Test0"
			}`,
			http.StatusOK,
		},
		{
			"/companies",
			`{
				"unknown": "Test0"
			}`,
			http.StatusOK,
		},
	}

	router := httprouter.New()
	router.POST("/companies", Create(model))

	for _, test := range tests {
		req, _ := http.NewRequest(http.MethodPost, test.url, strings.NewReader(test.body))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if status := rr.Code; status != test.want {
			t.Errorf("%v %v status %v; want status %v", req.Method, test.url, status, test.want)
		}
	}
}

func TestDelete(t *testing.T) {
	stmts := []string{
		"insert into company (id, name, code, country, website, phone) values (100, 'Test0', 'TST0', 'TT', 'NA', '0')",
	}
	pool, err := setup(stmts)
	if err != nil {
		t.Fatal(err)
	}
	defer clean(pool, []string{
		"delete from company where name in ('Test0')",
	})

	model := NewModel(pool)

	tests := []struct {
		url  string
		want int
	}{
		{
			"/companies/100",
			http.StatusOK,
		},
		{
			"/companies/200",
			http.StatusBadRequest,
		},
	}

	router := httprouter.New()
	router.DELETE("/companies/:id", Delete(model))

	for _, test := range tests {
		req, _ := http.NewRequest(http.MethodDelete, test.url, nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if status := rr.Code; status != test.want {
			t.Errorf("%v %v status %v; want status %v", req.Method, test.url, status, test.want)
		}
	}
}

func TestUpdate(t *testing.T) {
	stmts := []string{
		"insert into company (id, name, code, country, website, phone) values (100, 'Test0', 'TST0', 'TT', 'NA', '0')",
	}
	pool, err := setup(stmts)
	if err != nil {
		t.Fatal(err)
	}
	defer clean(pool, []string{
		"delete from company where name in ('Test0', 'Test1')",
	})

	model := NewModel(pool)

	tests := []struct {
		url  string
		body string
		want int
	}{
		{
			"/companies/100",
			`{
				"name": "Test1"
			}`,
			http.StatusOK,
		},
		{
			"/companies/200",
			`{
				"name": "Test1"
			}`,
			http.StatusBadRequest,
		},
	}

	router := httprouter.New()
	router.PUT("/companies/:id", Update(model))

	for _, test := range tests {
		req, _ := http.NewRequest(http.MethodPut, test.url, strings.NewReader(test.body))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if status := rr.Code; status != test.want {
			t.Errorf("%v %v status %v; want status %v", req.Method, test.url, status, test.want)
		}
	}
}

func TestBasicAuth(t *testing.T) {
	tests := []struct {
		url  string
		user string
		pass string
		want int
	}{
		{
			"/companies",
			"user",
			"pass",
			http.StatusOK,
		},
		{
			"/companies",
			"wrong",
			"wrong",
			http.StatusUnauthorized,
		},
	}

	var handler httprouter.Handle = func(w http.ResponseWriter, r *http.Request, p httprouter.Params) { w.WriteHeader(200) }

	router := httprouter.New()
	router.POST("/companies", BasicAuth(handler, "user", "pass"))

	for _, test := range tests {
		req, _ := http.NewRequest(http.MethodPost, test.url, nil)
		req.SetBasicAuth(test.user, test.pass)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if status := rr.Code; status != test.want {
			t.Errorf("%v %v status %v; want status %v", req.Method, test.url, status, test.want)
		}
	}
}

func TestCyprusRequest(t *testing.T) {
	tests := []struct {
		url        string
		remoteaddr string
		want       int
	}{
		{
			"/companies",
			"102.38.233.1:34343",
			http.StatusOK,
		},
		{
			"/companies",
			"8.8.8.8:34343",
			http.StatusUnauthorized,
		},
	}

	var handler httprouter.Handle = func(w http.ResponseWriter, r *http.Request, p httprouter.Params) { w.WriteHeader(200) }

	router := httprouter.New()
	router.POST("/companies", CyprusRequest(handler))

	for _, test := range tests {
		req, _ := http.NewRequest(http.MethodPost, test.url, nil)
		req.RemoteAddr = test.remoteaddr
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		bb, _ := io.ReadAll(rr.Body)

		if status := rr.Code; status != test.want {
			t.Errorf("%v %v status %v; want status %v; message %q", req.Method, test.url, status, test.want, string(bb))
		}
	}
}
