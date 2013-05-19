package checker

import (
    "time"
    "net/http"
    "net/url"
    "io/ioutil"
)

type site struct {
    url    *url.URL
    out    chan *CheckResult
    client *http.Client
    stop   chan bool
    delay  time.Duration
}

func newSite(u *url.URL, delay time.Duration, out chan *CheckResult) *site {
    s := new(site)

    s.client = &http.Client{CheckRedirect: checkRedirect}
    s.out = out
    s.stop = make(chan bool)
    s.url = u
    s.delay = delay

    return s
}

func (s *site) start() {

    log("Checking site %s every %d seconds\n", s.url.String(), s.delay)
    t := time.Tick(s.delay * time.Second)

    for {
        select {
            case _ = <-t:
                s.check()
            case _ = <-s.stop:
                return
        }
    }
}

func (s *site) check() {

    log("Getting %s \n", s.url.String())

    start := time.Now()
    resp, err := s.client.Get(s.url.String())
    connect_time := time.Now()

    if err != nil {
        log(err.Error())
        return
    }

    data, err := ioutil.ReadAll(resp.Body)
    rcv_time := time.Now()

    resp.Body.Close()

    if err != nil {
        log(err.Error())
        return
    }

    s.out <- &CheckResult{Resp: resp, Body: data, Connecting: connect_time.Sub(start), Receiving: rcv_time.Sub(connect_time), Timestamp : start, Url : s.url.String()}
}
