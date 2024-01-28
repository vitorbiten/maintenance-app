#!/bin/sh
chmod +rx /scripts/k6.js
exec k6 run /scripts/k6.js
