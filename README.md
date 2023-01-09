# Mandelbrot LoadBalancer
## API Documentation

### Endpoints

-   `/form` : Returns the HTML form for inputting Mandelbrot set parameters. This route takes no parameters.
    
-   `/mandelbrot` : Generates an image of the Mandelbrot set using the parameters provided in the POST request. The request parameters should be in the form of an HTML form and should include:
    
    -   `realMin`: Lower limit of the x-axis.
    -   `realMax`: Upper limit of the x-axis.
    -   `imagMin`: Lower limit of the y-axis.
    -   `imagMax`: Upper limit of the y-axis.
    -   `color`: Color value used to generate the image.

### Load Balancer Strategy

The server includes a load balancer that distributes requests between multiple simple servers (see below). Requests are sent to the load balancer via the API at `localhost:8000`. The load balancer uses a "round-robin" load balancing strategy, sending requests to each simple server in turn, starting with the first server in the list `localhost:8081`

### Libraries Used

-   `image`: Go library for manipulating images.
-   `image/color`: Go library for manipulating colors in images.
-   `image/png`: Go library for encoding and decoding PNG images.
-   `math/cmplx`: Go library for working with complex numbers.
-   `net/http`: Go library for handling HTTP requests.
-   `net/http/httputil`: Go library for helping to debug HTTP requests.
-   `net/url`: Go library for manipulating URLs.
-   `os`: Go library for interacting with the operating system.
-   `runtime`: Go library for getting information about the execution of the application.
-   `strconv`: Go library for converting strings to numerical values.
-   `sync`: Go library for synchronizing the execution of multiple goroutines.
