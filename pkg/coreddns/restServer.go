package coreddns

import (
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var updateCh chan<- Record

var ipv4Regex = regexp.MustCompile("^(\\d+)\\.(\\d+)\\.(\\d+)\\.(\\d+)$")
var successContent = []byte("SUCCESS")
var errorContent = []byte("ERROR")

type server struct {
	domain        string
	leaseDuration int64
	updater       *EtcdUpdater
}

func getIP(query url.Values, removeAddr string) string {
	ip := ""
	if _, ok := query["ip"]; ok {
		ip = query["ip"][0]
	} else {
		ip = removeAddr
		idx := strings.Index(removeAddr, ":")
		if idx > -1 {
			ip = removeAddr[0:idx]
		}
	}

	return ip
}

func (s *server) set(w http.ResponseWriter, r *http.Request, leaseSeconds int64) {
	u, _ := url.Parse(r.RequestURI)
	query := u.Query()

	ip := getIP(query, r.RemoteAddr)
	if !ipv4Regex.MatchString(ip) {
		w.WriteHeader(500)
		w.Write([]byte("IPV4-ONLY"))
		return
	}

	names, ok := query["name"]
	if !ok {
		w.WriteHeader(500)
		w.Write(errorContent)
		return
	}

	for i, name := range names {
		s.updater.SetARecord(name, ip, leaseSeconds)
		if i == 0 {
			s.updater.SetReverseRecord(name, ip, leaseSeconds)
		}
	}

	w.WriteHeader(200)
	w.Write(successContent)
}

func (s *server) indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(successContent)
}

func (s *server) setHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RequestURI)
	s.set(w, r, 0)
}

func (s *server) leaseHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RequestURI)
	s.set(w, r, s.leaseDuration)
}

func (s *server) unsetHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RequestURI)

	u, _ := url.Parse(r.RequestURI)
	query := u.Query()

	if _, ok := query["name"]; ok {
		for _, name := range query["name"] {
			s.updater.UnsetARecord(name)
		}
	}

	if _, ok := query["ip"]; ok {
		for _, ip := range query["ip"] {
			s.updater.UnsetReverseRecord(ip)
		}
	}

	w.WriteHeader(200)
	w.Write(successContent)
	return
}

func StartServer(servers []string, domain string, lease int64) {
	log.Println("* starting service")

	updater, err := NewUpdater(servers, domain)
	if err != nil {
		panic(err)
	}

	server := &server{
		domain:        domain,
		leaseDuration: lease,
		updater:       updater,
	}

	http.HandleFunc("/", server.indexHandler)
	http.HandleFunc("/set", server.setHandler)
	http.HandleFunc("/unset", server.unsetHandler)
	http.HandleFunc("/lease", server.leaseHandler)

	log.Println("  listening 12379")
	log.Fatal(http.ListenAndServe(":12379", nil))
}
