#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import requests

def print_request(request):
    req = "{method} {path_url} HTTP/1.1\r\n{headers}\r\n{body}".format(
        method = request.method,
        path_url = request.path_url,
        headers = ''.join('{0}: {1}\r\n'.format(k, v) for k, v in request.headers.items()),
        body = request.body or "",
    )
    return "{req_size}\n{req}\r\n".format(req_size = len(req), req = req)

#POST multipart form data
def post_multipart(host, port, namespace, headers, payload):
    req = requests.Request(
        'POST',
        'https://{host}:{port}{namespace}'.format(
            host = host,
            port = port,
            namespace = namespace,
        ),
        headers = headers,
        data = payload,
    )
    prepared = req.prepare()
    return print_request(prepared)

if __name__ == "__main__":
    host = 'localhost'
    port = '8080'
    namespace = '/api/auth'
    headers = {
        'Accept': 'application/json',
        'Content-type': 'application/json'
    }
    payload = {
        "username": "some_user1", "password": "pass1"
    }

    print(post_multipart(host, port, namespace, headers, payload))
