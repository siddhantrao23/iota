const http = require('http');
const vm = require('vm');

const server = http.createServer((req, res) => {
  if (req.method !== 'POST') {
    res.writeHead(405);
    return res.end(JSON.stringify({ output: '', error: 'Method not allowed' }));
  }

  let body = '';
  req.on('data', chunk => (body += chunk));
  req.on('end', () => {
    try {
      const data = JSON.parse(body);
      const args = data.args || {};
      const code = (typeof args === 'object' && args.code) || '';

      let output = '';
      const originalLog = console.log;
      console.log = (...args) => {
        output += args.map(a => (typeof a === 'object' ? JSON.stringify(a) : String(a))).join(' ') + '\n';
      };

      try {
        vm.runInNewContext(code, { console, setTimeout, clearTimeout, Buffer, require });
        console.log = originalLog;
        res.writeHead(200, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ output, error: '' }));
      } catch (e) {
        console.log = originalLog;
        res.writeHead(400, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ output: '', error: e.message }));
      }
    } catch (e) {
      res.writeHead(400, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify({ output: '', error: 'Invalid JSON' }));
    }
  });
});

const PORT = 8080;
server.listen(PORT, () => {
  console.log(`JavaScript worker started on port ${PORT}...`);
});
