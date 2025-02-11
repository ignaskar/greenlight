package main

import (
	"flag"
	"log"
	"net/http"
)

/*
This example requires a user to be created with following credentials

{
	"name": "test user",
	"email": "test@example.com",
	"password": "test1234"
}
*/

const html = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
</head>
<body>
	<h1>Simple CORS</h1>
	<div id="output"></div>
    <script>
        document.addEventListener('DOMContentLoaded', function() {
            fetch("http://localhost:4000/v1/tokens/authentication",
				{
					method: "POST",
					headers: {
						'Content-Type': 'application/json'
					},
					body: JSON.stringify({
						"email": "test@example.com",
						"password": "test1234"
					})
				})
                .then(
                    function(response) {
                        response.text().then(function(text) {
                            document.getElementById("output").innerHTML = text;
                        });
                    },
                    function(err) {
                        document.getElementById("output").innerHTML = err;
                    });
        });
    </script>
</body>
</html>
`

func main() {
	addr := flag.String("addr", ":9000", "Server address")
	flag.Parse()

	log.Printf("starting server on %s", *addr)

	err := http.ListenAndServe(*addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(html))
	}))
	log.Fatal(err)
}
