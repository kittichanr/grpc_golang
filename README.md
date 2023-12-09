# grpc_golang

### Setup Nginx

1. `brew install nginx`

2. Go to path `cd /usr/local/etc/nginx`

3. copy this config to the file nginx.conf

```

worker_processes  1;

error_log  /usr/local/var/log/nginx/error.log;

events {
    worker_connections  10;
}

http {
    access_log  /usr/local/var/log/nginx/access.log;

    upstream auth_services {
        server 0.0.0.0:50051;
    }

    upstream laptop_services {
        server 0.0.0.0:50052;
    }

    server {
        listen       8080 ssl;
        http2 on;

        ssl_certificate cert/server-cert.pem;
        ssl_certificate_key cert/server-key.pem;

        ssl_client_certificate cert/ca-cert.pem;
        ssl_verify_client on;

        location /pcbook.v1.AuthService {
            grpc_pass grpcs://auth_services;

            grpc_ssl_certificate cert/server-cert.pem;
            grpc_ssl_certificate_key cert/server-key.pem;
        }

         location /pcbook.v1.LaptopService {
            grpc_pass grpcs://laptop_services;

            grpc_ssl_certificate cert/server-cert.pem;
            grpc_ssl_certificate_key cert/server-key.pem;
        }
    }  
}

```

4. create folder cert `mkdir cert`

- copy server-cert.pem, server-key.pem, ca-cert.pem from this repo to cert folder in `/usr/local/etc/nginx/cert`

5. run nginx  `nginx`

- check nginx running `ps aux | grep nginx`

- if you want stop/restart use `nginx -s stop`

6. start server and client to test load balance with nginx 
- `make server1-tls` 
- `make server2-tls` 
- `make client-tls`