from http.server import BaseHTTPRequestHandler, HTTPServer
import json
import subprocess
import os
import urllib.request

class ShellHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        content_length = int(self.headers['Content-Length'])
        post_data = self.rfile.read(content_length)

        try:
            data = json.loads(post_data.decode('utf-8'))
            req_type = data.get('type', '')
            args = data.get('args', {})

            if req_type == 'shell':
                cmd = args.get('command', '')
                result = subprocess.run(cmd, shell=True, capture_output=True, text=True, timeout=30)
                output = result.stdout
                error = result.stderr
                if result.returncode != 0:
                    error = error or f'exit code {result.returncode}'

            elif req_type == 'read_file':
                path = args.get('path', '')
                try:
                    with open(path, 'r') as f:
                        output = f.read()
                    error = ''
                except Exception as e:
                    output = ''
                    error = str(e)

            elif req_type == 'write_file':
                path = args.get('path', '')
                content = args.get('content', '')
                mode = 'a' if args.get('append') else 'w'
                try:
                    with open(path, mode) as f:
                        f.write(content)
                    output = f'written {len(content)} bytes to {path}'
                    error = ''
                except Exception as e:
                    output = ''
                    error = str(e)

            elif req_type == 'list_dir':
                path = args.get('path', '.')
                try:
                    entries = os.listdir(path)
                    output = '\n'.join(entries)
                    error = ''
                except Exception as e:
                    output = ''
                    error = str(e)

            elif req_type == 'grep':
                pattern = args.get('pattern', '')
                path = args.get('path', '.')
                recursive = args.get('recursive', False)
                flag = '-r' if recursive else ''
                try:
                    result = subprocess.run(
                        f'grep {flag} "{pattern}" {path}',
                        shell=True, capture_output=True, text=True, timeout=30
                    )
                    output = result.stdout
                    error = result.stderr
                except Exception as e:
                    output = ''
                    error = str(e)

            elif req_type == 'http':
                method = args.get('method', 'GET').upper()
                url = args.get('url', '')
                body = args.get('body', '')
                headers = args.get('headers', {})
                try:
                    req = urllib.request.Request(url, data=body.encode() if body else None,
                                                 headers=headers, method=method)
                    with urllib.request.urlopen(req, timeout=30) as resp:
                        output = resp.read().decode()
                    error = ''
                except Exception as e:
                    output = ''
                    error = str(e)

            else:
                output = ''
                error = f'unknown type: {req_type}'

            status = 200 if not error else 400
            response = {"output": output, "error": error}

        except json.JSONDecodeError:
            status = 400
            response = {"output": "", "error": "Invalid JSON"}
        except subprocess.TimeoutExpired:
            status = 400
            response = {"output": "", "error": "command timed out"}

        self.send_response(status)
        self.send_header('Content-type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps(response).encode('utf-8'))

def run(server_class=HTTPServer, handler_class=ShellHandler, port=8080):
    server_address = ('', port)
    print(f"Shell worker started on port {port}...")
    httpd = server_class(server_address, handler_class)
    httpd.serve_forever()

if __name__ == "__main__":
    run()
