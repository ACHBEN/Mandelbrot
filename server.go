package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/cmplx"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"sync"
)

func main() {
	servers := []Server{}
	numServers := runtime.NumCPU()
	for i := 0; i < numServers; i++ {
		url := fmt.Sprintf("http://localhost:%d", 8081+i)
		servers = append(servers, newSimpleServer(url))
		go start(i+1, 8081+i)
	}
	lb := NewLoadBalancer("8000", servers)
	handleRedirect := func(rw http.ResponseWriter, req *http.Request) {
		lb.serveProxy(rw, req)
	}
	fmt.Printf("serving requests at 'localhost:%s'\n", lb.port)
	// register a proxy handler to handle all requests
	http.HandleFunc("/", handleRedirect)
	http.ListenAndServe(":"+lb.port, nil)

}
func start(id, port int) { // worker en gros les serveurs
	mux := http.NewServeMux()
	mux.HandleFunc("/form", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "formulaire.html") // afficher le formulaire html
	})
	mux.HandleFunc("/mandelbrot", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			fmt.Fprintf(w, "404 page not found")
		} else if r.Method == http.MethodPost {
			r.ParseForm()
			realMin, _ := strconv.ParseFloat(r.Form.Get("realMin"), 64)
			realMax, _ := strconv.ParseFloat(r.Form.Get("realMax"), 64)
			imagMin, _ := strconv.ParseFloat(r.Form.Get("imagMin"), 64)
			imagMax, _ := strconv.ParseFloat(r.Form.Get("imagMax"), 64)
			color, _ := strconv.ParseFloat(r.Form.Get("color"), 64)
			img := mandelbrot(color, realMin, realMax, imagMin, imagMax, 9000, 8000, 8000)
			w.Header().Set("Content-Type", "image/png")
			png.Encode(w, img)
		}
	})
	http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

func mandelbrot(color, realMin, realMax, imagMin, imagMax float64, width, height, iterations int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	wg := sync.WaitGroup{}
	for py := 0; py < height; py++ {
		wg.Add(1)
		go func(py int) { // worker on prend une ligne de l'image et un thread va calculer mandelbrot de chaque pixel de cette ligne
			y := float64(py)/float64(height)*(imagMax-imagMin) + imagMin
			for px := 0; px < width; px++ {
				x := float64(px)/float64(width)*(realMax-realMin) + realMin
				z := complex(x, y)
				img.Set(px, py, computeMandelbrot(z, iterations, color))
			}
			wg.Done()
		}(py)
	}
	wg.Wait()
	return img
}

func computeMandelbrot(z complex128, iterations int, couleur float64) color.Color {

	var v complex128
	for n := uint8(0); n < uint8(iterations); n++ {
		v = v*v + z
		if cmplx.Abs(v) > 2 {
			if couleur == 1 {
				return color.RGBA{11, 14, 129 * n, 200} ///color.RGBA{0, n, 0, 255} rgba(11, 14, 129, 0.92)
			}
			if couleur == 2 {
				return color.RGBA{0, 255 - 32*n, 8, 250} ///color.RGBA{0, n, 0, 255} rgba(11, 14, 129, 0.92)
			}
			if couleur == 3 {
				return color.Gray{255 - 32*n} ///color.RGBA{0, n, 0, 255} rgba(11, 14, 129, 0.92)
			}

		}
	}
	return color.Black
}

type LoadBalancer struct {
	port            string
	roundRobinCount int
	servers         []Server
}

func NewLoadBalancer(port string, servers []Server) *LoadBalancer {
	return &LoadBalancer{
		port:            port,
		roundRobinCount: 0,
		servers:         servers,
	}
}

type Server interface {
	// Address returns the address with which to access the server
	Address() string

	// IsAlive returns true if the server is alive and able to serve requests
	IsAlive() bool

	// Serve uses this server to process the request
	Serve(rw http.ResponseWriter, req *http.Request)
}

type simpleServer struct {
	addr  string
	proxy *httputil.ReverseProxy
}

func (s *simpleServer) Address() string { return s.addr }

func (s *simpleServer) IsAlive() bool { return true }

func (s *simpleServer) Serve(rw http.ResponseWriter, req *http.Request) {
	s.proxy.ServeHTTP(rw, req)
}

func handleErr(err error) {
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
func newSimpleServer(addr string) *simpleServer {
	serverUrl, err := url.Parse(addr)
	handleErr(err)

	return &simpleServer{
		addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

// getNextServerAddr returns the address of the next available server to send a
// request to, using a simple round-robin algorithm
func (lb *LoadBalancer) getNextAvailableServer() Server {
	server := lb.servers[lb.roundRobinCount%len(lb.servers)]
	for !server.IsAlive() {
		lb.roundRobinCount++
		server = lb.servers[lb.roundRobinCount%len(lb.servers)]
	}
	lb.roundRobinCount++

	return server
}

func (lb *LoadBalancer) serveProxy(rw http.ResponseWriter, req *http.Request) {
	targetServer := lb.getNextAvailableServer()

	// could optionally log stuff about the request here!
	fmt.Printf("forwarding request to address %q\n", targetServer.Address())

	// could delete pre-existing X-Forwarded-For header to prevent IP spoofing
	targetServer.Serve(rw, req)
}
