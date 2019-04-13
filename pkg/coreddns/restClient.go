package coreddns

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

var index = 0

func getUpdateServerUri(servers []string, cmd string, names []string, ip string) string {
	index++
	server := servers[index%len(servers)]
	uu := url.URL{
		Scheme: "http",
		Host:   server + ":12379",
		Path:   "/" + cmd,
	}

	query := uu.Query()
	for _, name := range names {
		query.Add("name", name)
	}
	if ip != "" {
		query.Add("ip", ip)
	}
	uu.RawQuery = query.Encode()

	return uu.String()
}

func send(uri string) error {
	resp, err := http.Get(uri)
	if err != nil {
		log.Println(err)
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println(string(body))
	return nil
}

func Set(servers []string, name string, ip string) error {
	uri := getUpdateServerUri(servers, "set", []string{name}, ip)
	return send(uri)
}

func Unset(servers []string, name string, ip string) error {
	uri := getUpdateServerUri(servers, "unset", []string{name}, ip)
	return send(uri)
}

func Lease(servers []string, names []string) error {
	uri := getUpdateServerUri(servers, "lease", names, "")
	return send(uri)
}
