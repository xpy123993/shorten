shorten is a lightweight URL shortener service.

In this implementation, all data will be stored in-memory, and dumped into the disk in JSON format.

Compile (Will store binaries under releases):
    ./release.sh

Usage:
    shorten_server --data=/path/to/link.json --addr=127.0.0.1:8080 \
                   --scheme-allowlist=http,https \
                   --update-path=/my/super/secret/path \
                   --prefix=example.com/link/

Under a reverse proxy (Nginx):
    server {
        ...
        location /link/ {
            proxy_pass http://127.0.0.1:8080/;
        }
    }

API:
- Directly visit http://your-address/<update-path>
- Rest API:
    
    $ curl -X POST -d "url=https://www.google.com/" -L "127.0.0.1:8080/my/super/secret/path"
    example.com/link/k-sC
    
    $ curl -L "127.0.0.1:8080/my/super/secret/path?url=https://www.google.com/"
    example.com/link/k-sC

