from http.server import HTTPServer, BaseHTTPRequestHandler
from datetime import datetime
import json

class HealthHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == '/health':
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            
            response = {
                'status': 'ok',
                'timestamp': datetime.utcnow().isoformat(),
                'service': 'user-service'
            }
            
            self.wfile.write(json.dumps(response).encode())
        else:
            self.send_response(404)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps({'error': 'Not found'}).encode())
    
    def log_message(self, format, *args):
        pass

if __name__ == '__main__':
    server = HTTPServer(('0.0.0.0', 8000), HealthHandler)
    print('User Service started on port 8000')
    server.serve_forever()