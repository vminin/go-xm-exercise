package company

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func List(model Model) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		companies, err := model.All(r.Form)
		if err != nil {
			status := http.StatusInternalServerError
			if _, ok := err.(ErrNotFound); ok {
				status = http.StatusBadRequest
			}
			http.Error(w, err.Error(), status)
			return
		}

		bytes, err := json.Marshal(companies)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	}
}

func Find(model Model) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		company, err := model.ByID(p.ByName("id"))
		if err != nil {
			status := http.StatusInternalServerError
			if _, ok := err.(ErrNotFound); ok {
				status = http.StatusBadRequest
			}
			http.Error(w, err.Error(), status)
			return
		}

		bytes, err := json.Marshal(company)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	}
}

func Create(model Model) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		c := new(Company)
		err := json.NewDecoder(r.Body).Decode(c)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		c, err = model.New(*c)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		bytes, err := json.Marshal(c)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	}
}

func Delete(model Model) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		company, err := model.ByID(p.ByName("id"))
		if err != nil {
			status := http.StatusInternalServerError
			if _, ok := err.(ErrNotFound); ok {
				status = http.StatusBadRequest
			}
			http.Error(w, err.Error(), status)
			return
		}

		if err := model.Delete(*company); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		bytes, err := json.Marshal(company)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	}
}

func Update(model Model) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		company, err := model.ByID(p.ByName("id"))
		if err != nil {
			status := http.StatusInternalServerError
			if _, ok := err.(ErrNotFound); ok {
				status = http.StatusBadRequest
			}
			http.Error(w, err.Error(), status)
			return
		}

		update := new(Company)
		err = json.NewDecoder(r.Body).Decode(update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		update.ID = company.ID

		if err := model.Update(*update); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		bytes, err := json.Marshal(update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	}
}

func BasicAuth(h httprouter.Handle, usr, pass string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		reqUsr, reqUsrPass, hasAuth := r.BasicAuth()
		if hasAuth && usr == reqUsr && pass == reqUsrPass {
			h(w, r, p)
		} else {
			w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}
	}
}

func CyprusRequest(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, fmt.Sprintf("remote address %q is not IP:port", r.RemoteAddr), http.StatusBadRequest)
			return
		}

		if userIP := net.ParseIP(ip); userIP == nil {
			http.Error(w, fmt.Sprintf("remote address %q is not IP:port", r.RemoteAddr), http.StatusBadRequest)
			return
		}

		cname, err := countryNameByIP(ip)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if cname == "Cyprus" {
			h(w, r, p)
		} else {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}
	}
}

func countryNameByIP(ip string) (string, error) {
	url := fmt.Sprintf("https://ipapi.co/%v/country_name", ip)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("GET %v: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET %v: %v", url, http.StatusText(resp.StatusCode))
	}

	bbytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("GET %v: %v", url, err)
	}

	return string(bbytes), nil
}
