import json
from http.server import BaseHTTPRequestHandler, HTTPServer


class RequestHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        content_length = int(self.headers['Content-Length'])
        post_data = self.rfile.read(content_length)

        data = json.loads(post_data)
        attachment = data['attachments'][0]

        print(f"\033[32m{attachment['title']}\033[0m\n{attachment['text']}")

        self.send_response(200)
        self.send_header('Content-type', 'application/json')
        self.end_headers()

        response = {'ok': True}
        self.wfile.write(json.dumps(response).encode('utf-8'))


def run(server_class=HTTPServer, handler_class=RequestHandler, port=8080):
    server_address = ('', port)
    httpd = server_class(server_address, handler_class)
    print(f"server listening at {port}...")
    httpd.serve_forever()


if __name__ == "__main__":
    run()
