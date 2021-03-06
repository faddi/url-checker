package checker

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
    stdlog "log"
    "os"
)

var curl string = "http://www.example.com"
var delay time.Duration = 5 *time.Second

func init(){

    l := stdlog.New(os.Stdout, "", stdlog.LstdFlags|stdlog.Lshortfile)

    SetLogger(l)
}

func failTimer(t *testing.T) {
    time.AfterFunc(10*time.Second, func() {
        t.Fatal("Test did not complete in 10 seconds")
    })
}

func Test_New(t *testing.T) {
    c := New()

    if c == nil {
        t.Error("nil returned from new")
    }

    if c.sites == nil {
        t.Error("urls not initiated")
    }
}

func Test_AddUrl(t *testing.T) {
    failTimer(t)
    c := New()

    if err := c.AddUrl("i am not a valid url", delay); err == nil {
        t.Error("Did not fail on invalid url")
    }

    if err := c.AddUrl("/ddfsd/das", delay); err == nil {
        t.Error("Did not fail on relative url")
    }

    if err := c.AddUrl(curl, delay); err != nil {
        t.Error("Failed to add proper url")
        t.Error(err.Error())
    }

    if _, ok := c.sites[curl]; ok != true {
        t.Error("A site should exist in the sites map if it is created without an error")
    }

}

func Test_StopUrl(t *testing.T) {
    failTimer(t)
    c := New()

    if err := c.AddUrl(curl, delay); err != nil {
        t.Error("Failed to add proper url")
        t.Error(err.Error())
    }

    time.Sleep(3 * time.Second)
    go func () {
        for {
            log("%v\n", <-c.ResultChan())
        }
    }()

    err := c.StopCheckingUrl(curl)

    if err != nil {
        t.Fatal(err.Error())
    }

    if _, ok := c.sites[curl]; ok == true {
        t.Error("A site should not exist in checker.sites after stop")
    }
}

func Test_Multiple(t *testing.T) {
    failTimer(t)
    c := New()

    urls := []string{curl, "http://www.google.com", "http://www.dn.se", "http://www.aftonbladet.se"}

    for _, u := range urls {
        if err := c.AddUrl(u, delay); err != nil {
            t.Error(err.Error())
        }
    }

    go func() {
        out := c.out
        for {
            d := <-out
            log("%s -> %s", d.Url, d.Resp.Status)

        }
    }()

    time.Sleep(5 * time.Second)

    for _, u := range urls {
        err := c.StopCheckingUrl(u)
        if err != nil {
            t.Fatal(err)
        }
    }

    if len(c.sites) > 0 {
        t.Error("No c.sites should be empty after all urls have been stopped")
    }

}

func Test_Redirect(t *testing.T) {
    failTimer(t)

    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
    }))
    defer ts.Close()

    c := New()

    err := c.AddUrl(ts.URL, delay)

    if err != nil {
        t.Error(err)
    }

    res := <-c.ResultChan()

    err = c.StopCheckingUrl(ts.URL)

    if err != nil {
        t.Error(err)
    }

    t.Log(res.Resp.Status)
}

func Test_404(t *testing.T) {
    failTimer(t)

    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        http.NotFound(w, r)
    }))
    defer ts.Close()

    c := New()

    err := c.AddUrl(ts.URL, delay)

    if err != nil {
        t.Error(err)
    }

    res := <-c.ResultChan()

    err = c.StopCheckingUrl(ts.URL)

    if err != nil {
        t.Error(err)
    }

    t.Log(res.Resp.Status)
    t.Log("%v", res.Resp)
}

func Test_NotExisting(t *testing.T) {
    //failTimer(t)

    c := New()

    url := "http://fakelocalurl:12345"
    //url := "http://www.google.com"

    err := c.AddUrl(url, delay)

    if err != nil {
        t.Error(err)
    }

    res := <-c.ResultChan()

    err = c.StopCheckingUrl(url)

    if err != nil {
        t.Error(err)
    }

    if res.Error == nil {
        t.Error("Response should have an error when getting an url that doesn't exist")
    }
}
