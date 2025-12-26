const assert = require('assert');
const fs = require('fs');

// Make sure there's a file like frontend/react/dist/assets/index-{hash}.js
// Vite uses base64-like hashes with alphanumeric chars and underscores
const files = fs.readdirSync('frontend/react/dist/assets');
console.log(files);
assert.ok(files.some((f) => /index-[A-Za-z0-9_-]+\.js/.test(f)));
