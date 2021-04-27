// GopherCon 2015: Tomas Senart - Embrace the Interface
// https://youtu.be/xyDkyFjzFVc

package main

import (
	"net/http"
	"time"
)

func main() {
	_ = Decorate(http.DefaultClient,
		Authorization("123"),
		FaultTolerance(5, time.Second),
	)
}

//Client sends http.Requests and returns http.Responses on errors in case if failure
type Client interface {
	Do(*http.Request) (*http.Response, error)
}

//ClientFunc is a function type that implements the Client interface
type ClientFunc func(r *http.Request) (res *http.Response, err error)

func (f ClientFunc) Do(r *http.Request) (*http.Response, error) {
	return f(r)
}

//Decorator wraps a Client with extra behaviour
type Decorator func(Client) Client

//Decorate decorated a Client c with all the given Decorators, in order
func Decorate(c Client, ds ...Decorator) Client {
	decorated := c
	for _, decorate := range ds {
		decorated = decorate(decorated)
	}
	return decorated
}

//FaultTolerance returns a Decorator that extends a Client with fault tolerance configured with given attempts and backoff duration.
func FaultTolerance(attempts int, backoff time.Duration) Decorator {
	return func(c Client) Client {
		return ClientFunc(func(r *http.Request) (res *http.Response, err error) {
			for i := 0; i <= attempts; i++ {
				if res, err = c.Do(r); err == nil {
					break
				}
				time.Sleep(backoff * time.Duration(i))
			}
			return res, err
		})
	}
}

//Authorization returns a Decorator that authorizes every Client request with the given token
func Authorization(token string) Decorator {
	return Header("Authorization", token)
}

//Header returns a Decorator that adds the given HTTP header to every request done by a Client
func Header(name, value string) Decorator {
	return func(c Client) Client {
		return ClientFunc(func(r *http.Request) (*http.Response, error) {
			r.Header.Add(name, value)
			return c.Do(r)
		})
	}
}
