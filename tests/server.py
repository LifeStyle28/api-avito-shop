import requests

from urllib.parse import urljoin

class Server:
    URL = "http://avito-shop-service:8080"

    def get(self, endpoint, headers):
        return requests.get(urljoin(self.URL, endpoint), headers=headers)

    def post(self, endpoint, data, headers=None):
        return requests.post(urljoin(self.URL, endpoint), json=data, headers=headers, verify=False)
