This Makefile uses openssl to generate certificates meant to be used for local testing.

Requirements:

OpenSSL.
GNU Make.

Usage:

In the root directory, type in 

```
make
```

Various SSL-related files and directories will be output at this point.

To start the test server, in the root directory type

```
node server/server.js
```

STEPS:
step_1: Make all the private keys
step_2: Make all the certs
step_3: Trust the root cert in the Mac keychain
step_4: Enable security.enterprise_roots.enabled, or make sure that you have an enterprise type security enabled for the browser of your choice.

...


Notes:
Only root server needs to be trusted.