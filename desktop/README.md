# Racore Desktop

The Electron shell runs the Racore UI in a hardened Chromium window and starts the Python `racored` service automatically. External sites use a persistent isolated Chromium session.

Development:

```powershell
npm run desktop:dev
```

Build the Windows installer:

```powershell
npm run desktop:package
```

The first packaged preview uses the private Racore hosted interface while all credentials, AI calls, IPFS operations, mesh identity, and domain keys stay in the local Python daemon.
